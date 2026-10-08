package httpapi

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientIPIgnoresSpoofedForwardHeaders(t *testing.T) {
	var got string
	h := clientIP(true)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.RemoteAddr }))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "1.2.3.4")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "10.0.0.5:1234" {
		t.Fatalf("forward headers must be ignored, got %q", got)
	}
	req.Header.Set("CF-Connecting-IP", "203.0.113.9")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !strings.HasPrefix(got, "203.0.113.9:") {
		t.Fatalf("CF-Connecting-IP should be used when trusted, got %q", got)
	}
	got = ""
	req.RemoteAddr = "10.0.0.5:1234"
	clientIP(false)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.RemoteAddr })).ServeHTTP(httptest.NewRecorder(), req)
	if got != "10.0.0.5:1234" {
		t.Fatalf("CF header must be ignored when not trusted, got %q", got)
	}
}

func TestAccessLogOmitsQueryString(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)
	h := accessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/update?token=SEGREDO&hostname=a.b.c", nil))
	if strings.Contains(buf.String(), "SEGREDO") || !strings.Contains(buf.String(), "/update") {
		t.Fatalf("log must keep the path and drop the query: %q", buf.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	securityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rec, httptest.NewRequest("GET", "/api/x", nil))
	for _, k := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if rec.Header().Get(k) == "" {
			t.Errorf("missing %s", k)
		}
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("API responses must not be cached")
	}
}
