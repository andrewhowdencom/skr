package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"testing"

	"github.com/andrewhowdencom/skr/pkg/store"
	"oras.land/oras-go/v2/content/oci"
)

// Helper to create a temporary OCI store
func createTestStore(t *testing.T) (string, *store.Store) {
	tempDir, err := os.MkdirTemp("", "skr-test-store")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	_, err = oci.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create OCI store: %v", err)
	}

	st, err := store.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	return tempDir, st
}

func TestServeOCIEndpoints(t *testing.T) {
	// Setup headers check helper
	checkHeaders := func(t *testing.T, w *httptest.ResponseRecorder) {
		if val := w.Header().Get("Docker-Distribution-API-Version"); val != "registry/2.0" {
			t.Errorf("Expected Docker-Distribution-API-Version header, got %s", val)
		}
		if val := w.Header().Get("Access-Control-Allow-Origin"); val != "*" {
			t.Errorf("Expected CORS header, got %s", val)
		}
	}

	// 1. Test _catalog (Empty)
	t.Run("Catalog Empty", func(t *testing.T) {
		tmpDir, st := createTestStore(t)
		defer os.RemoveAll(tmpDir)

		ctx := context.Background()
		handler := newOCIHandler(ctx, st, nil)

		req := httptest.NewRequest("GET", "/v2/_catalog", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}
		checkHeaders(t, w)

		var body map[string][]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(body["repositories"]) != 0 {
			t.Errorf("Expected empty repositories, got %v", body["repositories"])
		}
	})

	// 2. Test Version Check
	t.Run("Version Check", func(t *testing.T) {
		tmpDir, st := createTestStore(t)
		defer os.RemoveAll(tmpDir)

		ctx := context.Background()
		handler := newOCIHandler(ctx, st, nil)

		req := httptest.NewRequest("GET", "/v2/", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", w.Code)
		}
		checkHeaders(t, w)
	})

	// 3. Test Proxy Mode (Simulated)
	// We can pass a proxy that points to a test server
	t.Run("Proxy Mode", func(t *testing.T) {
		// Mock upstream
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusAccepted) // Custom code to verify proxy
		}))
		defer upstream.Close()

		targetURL, _ := url.Parse(upstream.URL)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		ctx := context.Background()
		handler := newOCIHandler(ctx, nil, proxy) // st is nil used in proxy mode

		req := httptest.NewRequest("GET", "/v2/_catalog", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected proxy to return 202 Accepted, got %d", w.Code)
		}
	})
}
