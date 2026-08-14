// Package proxy builds HTTP clients and Delta Chat proxy URLs for SOCKS/HTTP proxies.
package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	xproxy "golang.org/x/net/proxy"
)

// Config describes an optional outbound proxy (SOCKS5 or HTTP CONNECT).
//
// URL forms:
//
//	socks5://[user:pass@]host:port
//	socks5h://[user:pass@]host:port   (remote DNS; treated like socks5)
//	http://[user:pass@]host:port
//	https://[user:pass@]host:port     (HTTP CONNECT to proxy)
//
// Or set Type/Host/Port/Username/Password when URL is empty.
type Config struct {
	// Enabled turns the proxy on. nil means "auto": on when a URL can be resolved.
	Enabled *bool `yaml:"enabled"`

	// URL is a full proxy URL (preferred).
	URL string `yaml:"url"`

	// Type is socks5, socks5h, http, or https when building from Host/Port.
	Type string `yaml:"type"`

	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// IsEnabled reports whether a usable proxy is configured and allowed.
func (c Config) IsEnabled() bool {
	u := c.ResolvedURL()
	if u == "" {
		return false
	}
	if c.Enabled != nil {
		return *c.Enabled
	}
	return true
}

// ResolvedURL returns the proxy URL or empty if not configured.
func (c Config) ResolvedURL() string {
	raw := strings.TrimSpace(c.URL)
	if raw != "" {
		return normalizeProxyURL(raw)
	}
	host := strings.TrimSpace(c.Host)
	if host == "" {
		return ""
	}
	scheme := strings.ToLower(strings.TrimSpace(c.Type))
	if scheme == "" {
		scheme = "socks5"
	}
	switch scheme {
	case "socks5", "socks5h", "http", "https", "socks":
		if scheme == "socks" {
			scheme = "socks5"
		}
	default:
		scheme = "socks5"
	}
	port := c.Port
	if port == 0 {
		if scheme == "http" || scheme == "https" {
			port = 8080
		} else {
			port = 1080
		}
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	if c.Username != "" || c.Password != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u.String()
}

func normalizeProxyURL(raw string) string {
	// Allow host:port without scheme → socks5
	if !strings.Contains(raw, "://") {
		raw = "socks5://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Scheme == "socks" {
		u.Scheme = "socks5"
	}
	return u.String()
}

// Validate checks the resolved URL when enabled.
func (c Config) Validate() error {
	if !c.IsEnabled() {
		return nil
	}
	raw := c.ResolvedURL()
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("proxy url: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "socks5", "socks5h", "http", "https":
	default:
		return fmt.Errorf("unsupported proxy scheme %q (use socks5, socks5h, http, https)", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("proxy url missing host")
	}
	return nil
}

// HTTPClient returns an *http.Client that dials via the proxy when enabled.
// If disabled, returns a plain client. timeout applies to the client.
func HTTPClient(c Config, timeout time.Duration) (*http.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	if !c.IsEnabled() {
		return &http.Client{Timeout: timeout}, nil
	}
	return httpClientForURL(c.ResolvedURL(), timeout)
}

func httpClientForURL(raw string, timeout time.Duration) (*http.Client, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		tr := &http.Transport{
			Proxy:                 http.ProxyURL(u),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		return &http.Client{Timeout: timeout, Transport: tr}, nil

	case "socks5", "socks5h":
		// golang.org/x/net/proxy: socks5 uses local DNS; socks5h remote DNS.
		// FromURL treats both.
		auth := &xproxy.Auth{}
		var authPtr *xproxy.Auth
		if u.User != nil {
			auth.User = u.User.Username()
			auth.Password, _ = u.User.Password()
			authPtr = auth
		}
		// Rebuild without user for dialer host
		host := u.Host
		var dialer xproxy.Dialer
		if authPtr != nil {
			dialer, err = xproxy.SOCKS5("tcp", host, authPtr, xproxy.Direct)
		} else {
			dialer, err = xproxy.SOCKS5("tcp", host, nil, xproxy.Direct)
		}
		if err != nil {
			return nil, fmt.Errorf("socks5 dialer: %w", err)
		}
		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Prefer DialContext if available
				if xd, ok := dialer.(xproxy.ContextDialer); ok {
					return xd.DialContext(ctx, network, addr)
				}
				return dialer.Dial(network, addr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		return &http.Client{Timeout: timeout, Transport: tr}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q", u.Scheme)
	}
}

// Merge returns child if it has a URL or explicit settings, else parent.
func Merge(parent, child Config) Config {
	out := parent
	if child.URL != "" || child.Host != "" || child.Type != "" {
		out = child
		// keep explicit enabled from child if set
		if child.Enabled != nil {
			out.Enabled = child.Enabled
		} else if parent.Enabled != nil && child.URL == "" && child.Host == "" {
			out.Enabled = parent.Enabled
		}
	} else if child.Enabled != nil {
		out.Enabled = child.Enabled
	}
	if child.Username != "" {
		out.Username = child.Username
	}
	if child.Password != "" {
		out.Password = child.Password
	}
	if child.Port != 0 {
		out.Port = child.Port
	}
	return out
}
