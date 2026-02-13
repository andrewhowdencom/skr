package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/instrumentation"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/andrewhowdencom/skr/pkg/ui"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var port int
var ociEndpoint string
var traceProvider string
var traceEndpoint string

var httpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Skills Registry UI locally",
	Long:  `Start a local HTTP server that builds and serves the UI on-the-fly, reflecting the current state of the OCI store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Mutually exclusive flags check
		if ociEndpoint != "" && ociPath != "" {
			return fmt.Errorf("cannot specify both --oci-path and --oci-endpoint")
		}

		// 0. Initialize Tracing
		shutdown, err := instrumentation.InitTracer(ctx, "skr", traceProvider, traceEndpoint)
		if err != nil {
			return fmt.Errorf("failed to initialize tracer: %w", err)
		}
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				fmt.Printf("Error shutting down tracer: %v\n", err)
			}
		}()

		var st *store.Store
		var proxy *httputil.ReverseProxy

		// 1. Initialize Backend (Store or Proxy)
		if ociEndpoint != "" {
			targetURL, err := url.Parse(ociEndpoint)
			if err != nil {
				return fmt.Errorf("invalid OCI endpoint URL: %w", err)
			}
			proxy = httputil.NewSingleHostReverseProxy(targetURL)
			// Update the director to set the host header to the target
			originalDirector := proxy.Director
			proxy.Director = func(req *http.Request) {
				originalDirector(req)
				req.Host = targetURL.Host
			}
			// Instrument Proxy Transport
			proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

			fmt.Printf("Mode: Remote Proxy to %s\n", ociEndpoint)
		} else {
			st, err = store.New(ociPath)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}
			if ociPath == "" {
				fmt.Println("Mode: Local System Store")
			} else {
				fmt.Printf("Mode: Local Store at %s\n", ociPath)
			}
		}

		// 2. Asset Handler
		assets, err := ui.Assets()
		if err != nil {
			return fmt.Errorf("failed to load embedded assets: %w", err)
		}

		// Use a local ServeMux to avoid global state issues
		mux := http.NewServeMux()
		fileServer := http.FileServer(http.FS(assets))

		// 4. OCI Registry Handlers
		mux.HandleFunc("/v2/", newOCIHandler(ctx, st, proxy))

		// 6. UI Assets
		mux.Handle("/", fileServer)

		// 7. Instrumentation
		handler := otelhttp.NewHandler(mux, "server")

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Serving Skills Registry (HTTP) at http://localhost%s\n", addr)
		fmt.Println("Press Ctrl+C to stop")

		if err := http.ListenAndServe(addr, handler); err != nil {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	},
}

func init() {
	httpCmd.AddCommand(httpServeCmd)
	httpServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	httpServeCmd.Flags().StringVar(&ociEndpoint, "oci-endpoint", "", "Remote OCI Registry endpoint to proxy (e.g. https://registry-1.docker.io)")
	httpServeCmd.Flags().StringVar(&traceProvider, "trace-provider", "none", "Trace provider to use (stdout, otlp, none)")
	httpServeCmd.Flags().StringVar(&traceEndpoint, "trace-endpoint", "", "Endpoint for the OTLP trace provider (e.g. localhost:4318)")
}

// newOCIHandler creates a handler for OCI registry endpoints
func newOCIHandler(GlobalCtx context.Context, st *store.Store, proxy *httputil.ReverseProxy) http.HandlerFunc {
	tracer := otel.Tracer("skr-oci-handler")

	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers for UI
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// PROXY MODE
		if proxy != nil {
			proxy.ServeHTTP(w, r)
			return
		}

		// LOCAL MODE
		// Determine span name
		path := strings.TrimPrefix(r.URL.Path, "/v2/")
		spanName := "oci.unknown"
		if path == "" {
			spanName = "oci.base"
		} else if path == "_catalog" {
			spanName = "oci.catalog"
		} else if strings.HasSuffix(path, "/tags/list") {
			spanName = "oci.tags.list"
		} else if strings.Contains(path, "/blobs/") {
			spanName = "oci.blob.fetch"
		} else if strings.Contains(path, "/manifests/") {
			spanName = "oci.manifest.fetch"
		}

		// Start Span
		ctx, span := tracer.Start(r.Context(), spanName)
		defer span.End()

		w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")

		// 4.1 API Version Check
		if path == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 4.2 Extensions: Catalog
		if path == "_catalog" {
			tags, err := st.List(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list tags: %v", err), http.StatusInternalServerError)
				return
			}

			repoSet := make(map[string]bool)
			for _, t := range tags {
				// repo:tag
				if lastIdx := strings.LastIndex(t, ":"); lastIdx != -1 {
					repoSet[t[:lastIdx]] = true
				} else {
					repoSet[t] = true
				}
			}

			var repos []string
			for repo := range repoSet {
				repos = append(repos, repo)
			}
			sort.Strings(repos)

			response := map[string][]string{"repositories": repos}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// 4.3 Tags List: <name>/tags/list
		if strings.HasSuffix(path, "/tags/list") {
			name := strings.TrimSuffix(path, "/tags/list")

			tags, err := st.List(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list tags: %v", err), http.StatusInternalServerError)
				return
			}

			var repoTags []string
			foundRepo := false

			prefix := name + ":"
			for _, t := range tags {
				if strings.HasPrefix(t, prefix) {
					foundRepo = true
					repoTags = append(repoTags, strings.TrimPrefix(t, prefix))
				} else if t == name {
					foundRepo = true
				}
			}

			if !foundRepo {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			sort.Strings(repoTags)

			response := map[string]interface{}{
				"name": name,
				"tags": repoTags,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// 4.4 Blobs: <name>/blobs/<digest>
		if idx := strings.LastIndex(path, "/blobs/"); idx != -1 {
			digestStr := path[idx+len("/blobs/"):]

			d := v1.Descriptor{
				Digest: digest.Digest(digestStr),
			}

			rc, err := st.Fetch(ctx, d)
			if err != nil {
				http.Error(w, fmt.Sprintf("Blob not found: %s", digestStr), http.StatusNotFound)
				return
			}
			defer rc.Close()

			if _, err := io.Copy(w, rc); err != nil {
				fmt.Printf("Error serving blob: %v\n", err)
			}
			return
		} else if idx := strings.LastIndex(path, "/manifests/"); idx != -1 {
			// 4.5 Manifests: <name>/manifests/<reference>
			name := path[:idx]
			ref := path[idx+len("/manifests/"):]

			var desc v1.Descriptor
			var err error

			if strings.HasPrefix(ref, "sha256:") {
				// Reference is a digest
				desc = v1.Descriptor{
					Digest: digest.Digest(ref),
				}
				if resolved, err := st.Resolve(ctx, ref); err == nil {
					desc = resolved
				}
			} else {
				// Reference is a tag
				fullRef := name + ":" + ref
				desc, err = st.Resolve(ctx, fullRef)
			}

			if err != nil {
				http.Error(w, fmt.Sprintf("Manifest not found: %s", ref), http.StatusNotFound)
				return
			}

			rc, err := st.Fetch(ctx, desc)
			if err != nil {
				http.Error(w, "Failed to fetch manifest content", http.StatusInternalServerError)
				return
			}
			defer rc.Close()

			w.Header().Set("Content-Type", desc.MediaType)
			w.Header().Set("Docker-Content-Digest", desc.Digest.String())
			w.Header().Set("Content-Length", fmt.Sprintf("%d", desc.Size))

			if _, err := io.Copy(w, rc); err != nil {
				fmt.Printf("Error serving manifest: %v\n", err)
			}
			return
		}

		http.NotFound(w, r)
	}
}
