package unit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/homealias/homealias/backend/internal/clientfiles"
	"github.com/homealias/homealias/backend/internal/http/update"
)

func TestClientFilesServeTemplates(t *testing.T) {
	for name, wantType := range map[string]string{"update.sh": "#!/bin/sh", "update.ps1": "param("} {
		h := clientfiles.Handler(func(*http.Request) string { return name })
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest("GET", "/client/"+name, nil))
		body := rec.Body.String()
		if rec.Code != 200 || !strings.Contains(body, wantType) {
			t.Fatalf("%s: code %d", name, rec.Code)
		}
		for _, p := range []string{"__HOMEALIAS_URL__", "__HOMEALIAS_TOKEN__", "__HOMEALIAS_HOSTNAME__"} {
			if !strings.Contains(body, p) {
				t.Errorf("%s sem placeholder %s", name, p)
			}
		}
		if !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") {
			t.Errorf("%s deveria ser anexo", name)
		}
	}
	rec := httptest.NewRecorder()
	clientfiles.Handler(func(*http.Request) string { return "../etc/passwd" })(rec, httptest.NewRequest("GET", "/client/x", nil))
	if rec.Code != 404 {
		t.Fatalf("arquivo desconhecido deve dar 404, deu %d", rec.Code)
	}
}

func TestDetectClientType(t *testing.T) {
	cases := map[string]string{
		"homealias-shell/1.0": "shell", "homealias-docker/1.0": "docker", "homealias-windows/1.0": "windows",
		"homealias-agent/0.1": "agent", "curl/8.0": "duckdns",
	}
	for ua, want := range cases {
		r := httptest.NewRequest("GET", "/update", nil)
		r.Header.Set("User-Agent", ua)
		if got := update.DetectClientType(r, "/update"); got != want {
			t.Errorf("%s -> %s, quero %s", ua, got, want)
		}
	}
	r := httptest.NewRequest("GET", "/nic/update", nil)
	r.Header.Set("User-Agent", "Mozilla")
	if update.DetectClientType(r, "/nic/update") != "dyndns" {
		t.Error("/nic/ deve ser dyndns")
	}
}
