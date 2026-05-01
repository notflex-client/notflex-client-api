package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// googleResolver resolves hostnames via Google's public DNS (8.8.8.8),
// bypassing the local router's DNS which may block certain domains.
var googleResolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "udp", "8.8.8.8:53")
	},
}

// proxyClient connects via Google DNS to bypass local DNS blocks.
var proxyClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			// Resolve through 8.8.8.8 instead of the local router DNS.
			ips, err := googleResolver.LookupHost(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("dns(8.8.8.8) lookup %s: %w", host, err)
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses for %s", host)
			}
			return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0], port))
		},
	},
}

// ProxyHLS proxies HLS m3u8 manifests and TS segments, rewriting segment
// URLs so they also pass through this proxy.
//
// GET /proxy/hls?url=<encoded_url>
func ProxyHLS(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "url query param required", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Referer", "https://phim.nguonc.com/")

	resp, err := proxyClient.Do(req)
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	content := string(body)
	contentType := resp.Header.Get("Content-Type")

	if isM3U8(rawURL, contentType, content) {
		content = rewriteM3U8(content, rawURL, r)
		contentType = "application/vnd.apple.mpegurl"
	} else if contentType == "" {
		contentType = "video/MP2T"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(content))
}

func isM3U8(rawURL, contentType, body string) bool {
	if strings.Contains(contentType, "mpegurl") {
		return true
	}
	if strings.HasSuffix(strings.Split(rawURL, "?")[0], ".m3u8") {
		return true
	}
	return strings.HasPrefix(strings.TrimSpace(body), "#EXTM3U")
}

func rewriteM3U8(content, manifestURL string, r *http.Request) string {
	parsed, err := url.Parse(manifestURL)
	if err != nil {
		return content
	}
	baseDir := parsed.Scheme + "://" + parsed.Host + path.Dir(parsed.Path) + "/"

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	proxyBase := scheme + "://" + r.Host + "/proxy/hls?url="

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		var absURL string
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			absURL = trimmed
		} else {
			absURL = baseDir + trimmed
		}
		lines[i] = proxyBase + url.QueryEscape(absURL)
	}
	return strings.Join(lines, "\n")
}
