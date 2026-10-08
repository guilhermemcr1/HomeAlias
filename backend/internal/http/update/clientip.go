package update

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request, trustCF bool) (net.IP, bool) {
	if !trustCF {
		return nil, false
	}
	raw := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
	if raw == "" {
		return nil, false
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return nil, false
	}
	return ip, true
}

func HasClientIPParams(r *http.Request) bool {
	q := r.URL.Query()
	if q.Get("ip") != "" || q.Get("myip") != "" {
		return true
	}
	_ = r.ParseForm()
	return r.Form.Get("ip") != "" || r.Form.Get("myip") != ""
}
