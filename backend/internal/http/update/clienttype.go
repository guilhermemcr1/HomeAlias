package update

import (
	"net/http"
	"strings"
)

func DetectClientType(r *http.Request, path string) string {
	ua := strings.ToLower(r.UserAgent())
	switch {
	case strings.Contains(ua, "homealias-agent"):
		return "agent"
	case strings.Contains(ua, "homealias-windows"):
		return "windows"
	case strings.Contains(ua, "homealias-shell"):
		return "shell"
	case strings.Contains(ua, "homealias-docker"):
		return "docker"
	case strings.HasPrefix(path, "/nic/"):
		return "dyndns"
	default:
		return "duckdns"
	}
}
