package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.4.2", "0.4.1", true},
		{"v0.4.2", "0.4.2", false},
		{"0.4.1", "0.4.2", false},
		{"1.0.0", "0.99.99", true},
		{"1.0.0", "1.0.0-beta.1", true},
		{"", "0.4.1", false},
	}
	for _, tt := range tests {
		if got := IsNewer(tt.latest, tt.current); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestCheckerReturnsLatestNPMVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"0.4.2"}`))
	}))
	defer server.Close()

	checker := Checker{URL: server.URL + "/latest", Client: server.Client()}
	result, err := checker.Check(context.Background(), "0.4.1")
	if err != nil {
		t.Fatal(err)
	}
	if result.LatestVersion != "0.4.2" || !result.UpdateAvailable {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCheckerRejectsFailedRegistryResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	checker := Checker{URL: server.URL, Client: server.Client()}
	if _, err := checker.Check(context.Background(), "0.4.1"); err == nil {
		t.Fatal("expected registry error")
	}
}
