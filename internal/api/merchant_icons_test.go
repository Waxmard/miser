package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Waxmard/miser/internal/repository/sqlite"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	repo, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("sqlite.New(:memory:) error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("repo.Migrate() error: %v", err)
	}

	server := httptest.NewServer(New(repo, nil).Handler())
	t.Cleanup(server.Close)
	return server
}

func TestMerchantIconsDeleteWithSlashInName(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()

	putReq, err := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/merchant-icons",
		strings.NewReader(`{"merchant_name":"FOO/BAR","icon_slug":"spotify"}`),
	)
	if err != nil {
		t.Fatalf("http.NewRequest(PUT) error: %v", err)
	}
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := client.Do(putReq)
	if err != nil {
		t.Fatalf("PUT /api/merchant-icons error: %v", err)
	}
	defer func() { _ = putResp.Body.Close() }()
	if putResp.StatusCode != http.StatusOK {
		t.Fatalf("PUT /api/merchant-icons status = %d, want %d", putResp.StatusCode, http.StatusOK)
	}

	assertMerchantIconNames(t, client, server.URL+"/api/merchant-icons", []string{"foo/bar"})

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/merchant-icons?name=FOO%2FBAR", http.NoBody)
	if err != nil {
		t.Fatalf("http.NewRequest(DELETE) error: %v", err)
	}
	deleteResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /api/merchant-icons error: %v", err)
	}
	defer func() { _ = deleteResp.Body.Close() }()
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE /api/merchant-icons status = %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
	}

	assertMerchantIconNames(t, client, server.URL+"/api/merchant-icons", nil)
}

func TestMerchantIconsDeleteRequiresName(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()

	for _, rawURL := range []string{
		server.URL + "/api/merchant-icons",
		server.URL + "/api/merchant-icons?name=",
		server.URL + "/api/merchant-icons?name=%20%20%20",
	} {
		req, err := http.NewRequest(http.MethodDelete, rawURL, http.NoBody)
		if err != nil {
			t.Fatalf("http.NewRequest(DELETE) error: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("DELETE %s error: %v", rawURL, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("DELETE %s status = %d, want %d", rawURL, resp.StatusCode, http.StatusBadRequest)
		}
	}
}

func assertMerchantIconNames(t *testing.T, client *http.Client, url string, want []string) {
	t.Helper()

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET /api/merchant-icons error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/merchant-icons status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var icons []merchantIconResponse
	if err := json.NewDecoder(resp.Body).Decode(&icons); err != nil {
		t.Fatalf("decode merchant icon response error: %v", err)
	}

	if len(icons) != len(want) {
		t.Fatalf("len(icons) = %d, want %d", len(icons), len(want))
	}

	for i, icon := range icons {
		if icon.MerchantName != want[i] {
			t.Fatalf("icons[%d].MerchantName = %q, want %q", i, icon.MerchantName, want[i])
		}
	}
}
