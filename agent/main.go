package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	base := os.Getenv("HOMEALIAS_URL")
	token := os.Getenv("HOMEALIAS_TOKEN")
	if base == "" || token == "" {
		log.Fatal("HOMEALIAS_URL and HOMEALIAS_TOKEN required")
	}
	interval := 5 * time.Minute
	if v := os.Getenv("HOMEALIAS_INTERVAL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			interval = time.Duration(n) * time.Second
		}
	}
	client := &http.Client{Timeout: 20 * time.Second}
	backoff := time.Second
	for {
		err := updateOnce(client, base, token)
		if err != nil {
			log.Printf("update error: %v", err)
			time.Sleep(backoff)
			if backoff < 5*time.Minute {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
		time.Sleep(interval)
	}
}

func updateOnce(client *http.Client, base, token string) error {
	// Dual-stack: prefer separate dials when possible; MVP sends one request using default stack.
	url := fmt.Sprintf("%s/update?token=%s", trimSlash(base), token)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "homealias-agent/0.1")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	// Optional second family attempt via forced network if both available
	if hasBothFamilies() {
		// Best-effort second call; server updates only the family of connecting IP.
		req2, _ := http.NewRequest(http.MethodGet, url, nil)
		req2.Header.Set("User-Agent", "homealias-agent/0.1")
		res2, err := client.Do(req2)
		if err == nil {
			res2.Body.Close()
		}
	}
	return nil
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func hasBothFamilies() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	var v4, v6 bool
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if ipnet.IP.To4() != nil {
				v4 = true
			} else if ipnet.IP.To16() != nil {
				v6 = true
			}
		}
	}
	return v4 && v6
}
