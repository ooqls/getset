package v1

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingServerStdImpl_GetLoggingConfig(t *testing.T) {
	srv := httptest.NewServer(Std())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/logging/config")
	if err != nil {
		t.Fatalf("Failed to get logging config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %d", resp.StatusCode)
	}

}

func TestLoggingServerStdImpl_SetLoggingConfig(t *testing.T) {
	srv := httptest.NewServer(Std())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/logging/config", "application/json", bytes.NewBuffer([]byte(`{"level": "DEBUG", "format": "JSON"}`)))
	if err != nil {
		t.Fatalf("Failed to set logging config: %v", err)
	}
	defer resp.Body.Close()
}

func TestLoggingServerStdImpl_UpdateLoggingLevel(t *testing.T) {
	srv := httptest.NewServer(Std())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/logging/level/DEBUG", "application/json", bytes.NewBuffer([]byte(`{"level": "DEBUG"}`)))
	if err != nil {
		t.Fatalf("Failed to update logging level: %v", err)
	}
	defer resp.Body.Close()
}

func TestLoggingServerStdImpl_UpdateLoggingFormat(t *testing.T) {
	srv := httptest.NewServer(Std())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/logging/format/JSON", "application/json", bytes.NewBuffer([]byte(`{"format": "JSON"}`)))
	if err != nil {
		t.Fatalf("Failed to update logging format: %v", err)
	}
	defer resp.Body.Close()
}
