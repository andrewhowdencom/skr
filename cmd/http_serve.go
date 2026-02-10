package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/andrewhowdencom/skr/pkg/ui"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var port int

var httpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Skills Registry UI locally",
	Long:  `Start a local HTTP server that builds and serves the UI on-the-fly, reflecting the current state of the OCI store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// 1. Initialize Store
		st, err := store.New(ociPath)
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 2. Asset Handler
		assets, err := ui.Assets()
		if err != nil {
			return fmt.Errorf("failed to load embedded assets: %w", err)
		}

		// Use a local ServeMux to avoid global state issues
		mux := http.NewServeMux()
		fileServer := http.FileServer(http.FS(assets))

		// 3. API Handler
		mux.HandleFunc("/api/skills", func(w http.ResponseWriter, r *http.Request) {
			// Scan OCI store on every request
			tags, err := st.List(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list skills: %v", err), http.StatusInternalServerError)
				return
			}

			type SkillVersion struct {
				Version string `json:"version"`
				Tag     string `json:"tag"`
			}

			type UISkill struct {
				ID          string         `json:"id"`   // Full repo name (e.g. ghcr.io/user/skill)
				Name        string         `json:"name"` // Short name (e.g. skill)
				Description string         `json:"description"`
				Author      string         `json:"author"`
				Versions    []SkillVersion `json:"versions"`
				LatestTag   string         `json:"latestTag"`
			}

			skillMap := make(map[string]*UISkill)

			for _, tag := range tags {
				// Parse Tag: repo:version
				repo := tag
				version := "latest"
				if lastIdx := lastIndex(tag, ":"); lastIdx != -1 {
					repo = tag[:lastIdx]
					version = tag[lastIdx+1:]
				}

				// Resolve to get metadata
				desc, err := st.Resolve(ctx, tag)
				if err != nil {
					continue
				}

				rc, err := st.Fetch(ctx, desc)
				if err != nil {
					continue
				}
				containerBytes, _ := io.ReadAll(rc)
				rc.Close()

				var manifest v1.Manifest
				json.Unmarshal(containerBytes, &manifest)

				// Get or Create Skill Entry
				if _, exists := skillMap[repo]; !exists {
					// Derive simplified name
					shortName := repo
					if lastSlash := lastIndex(repo, "/"); lastSlash != -1 {
						shortName = repo[lastSlash+1:]
					}

					skillMap[repo] = &UISkill{
						ID:          repo,
						Name:        shortName,
						Description: manifest.Annotations["com.skr.description"],
						Author:      manifest.Annotations["com.skr.author"],
						Versions:    []SkillVersion{},
					}
				}

				// Append Version
				displayVersion := version
				if v := manifest.Annotations["com.skr.version"]; v != "" {
					displayVersion = v
				}

				skillMap[repo].Versions = append(skillMap[repo].Versions, SkillVersion{
					Version: displayVersion,
					Tag:     tag,
				})

				// Update metadata if missing
				if skillMap[repo].Description == "" {
					skillMap[repo].Description = manifest.Annotations["com.skr.description"]
				}
				if skillMap[repo].Author == "" {
					skillMap[repo].Author = manifest.Annotations["com.skr.author"]
				}
			}

			// Convert Map to Slice
			var uiSkills []UISkill
			for _, s := range skillMap {
				uiSkills = append(uiSkills, *s)
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(uiSkills); err != nil {
				http.Error(w, "Failed to encode skills", http.StatusInternalServerError)
			}
		})

		// 4. OCI Registry Handlers (Manual Routing)
		mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
			// Always set API Version Header
			w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")

			path := strings.TrimPrefix(r.URL.Path, "/v2/")

			// 4.1 API Version Check
			if path == "" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Parse path for blobs or manifests
			// Format: <name>/blobs/<digest> OR <name>/manifests/<reference>
			// We look for the LAST occurrence of "/blobs/" or "/manifests/" to split

			if idx := strings.LastIndex(path, "/blobs/"); idx != -1 {
				// BLOB FETCH
				// name := path[:idx] (unused currently as content is CAS)
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
				// MANIFEST FETCH
				name := path[:idx]
				ref := path[idx+len("/manifests/"):]

				var desc v1.Descriptor
				var err error

				if strings.HasPrefix(ref, "sha256:") {
					// Reference is a digest
					// Try to resolve using digest itself if supported, or rely on fetch
					desc = v1.Descriptor{
						Digest: digest.Digest(ref),
					}
					// Attempt resolve to get MediaType
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
		})

		// 5. UI Redirect
		mux.HandleFunc("/skills.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/api/skills", http.StatusTemporaryRedirect)
		})

		// 6. UI Assets
		mux.Handle("/", fileServer)

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Serving Skills Registry (HTTP) at http://localhost%s\n", addr)
		if ociPath == "" {
			fmt.Println("Source: System OCI Store")
		} else {
			fmt.Printf("Source: %s\n", ociPath)
		}
		fmt.Println("Press Ctrl+C to stop")

		if err := http.ListenAndServe(addr, mux); err != nil {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	},
}

func init() {
	httpCmd.AddCommand(httpServeCmd)
	httpServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
}
