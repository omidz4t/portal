package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func boolPtr(b bool) *bool { return &b }

func TestResolvedURL(t *testing.T) {
	c := Config{URL: "socks5://127.0.0.1:1080"}
	if got := c.ResolvedURL(); got != "socks5://127.0.0.1:1080" {
		t.Fatalf("url: %s", got)
	}
	c = Config{Host: "10.0.0.1", Port: 9050, Type: "socks5"}
	if got := c.ResolvedURL(); got != "socks5://10.0.0.1:9050" {
		t.Fatalf("built: %s", got)
	}
	c = Config{URL: "127.0.0.1:1080"}
	if got := c.ResolvedURL(); got != "socks5://127.0.0.1:1080" {
		t.Fatalf("normalize: %s", got)
	}
}

func TestIsEnabled(t *testing.T) {
	if (Config{}).IsEnabled() {
		t.Fatal("empty should be disabled")
	}
	c := Config{URL: "socks5://127.0.0.1:1", Enabled: boolPtr(false)}
	if c.IsEnabled() {
		t.Fatal("explicit false")
	}
	c = Config{URL: "socks5://127.0.0.1:1"}
	if !c.IsEnabled() {
		t.Fatal("url implies enabled")
	}
}

// TestHTTPProxyTransparentConnect verifies HTTP CONNECT-style proxying:
// client → HTTP proxy → origin HTTP server.
func TestHTTPProxyTransparentConnect(t *testing.T) {
	var originHits atomic.Int32
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHits.Add(1)
		_, _ = w.Write([]byte("pong"))
	}))
	defer origin.Close()

	// Reverse proxy that forwards absolute-form requests (http proxy style).
	var proxyHits atomic.Int32
	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHits.Add(1)
		// Transparent forward: dial origin as requested URL or Host.
		target := r.URL.String()
		if !r.URL.IsAbs() {
			target = "http://" + r.Host + r.URL.RequestURI()
		}
		req, err := http.NewRequest(r.Method, target, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
	defer proxySrv.Close()

	client, err := HTTPClient(Config{URL: proxySrv.URL}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Get(origin.URL + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Fatalf("body %q", body)
	}
	if proxyHits.Load() < 1 {
		t.Fatal("proxy was not used")
	}
	if originHits.Load() < 1 {
		t.Fatal("origin was not hit")
	}
}

// TestSOCKS5ProxyDial verifies a real SOCKS5 handshake (no auth) and that
// HTTP traffic reaches the origin only via the SOCKS server (transparent tunnel).
func TestSOCKS5ProxyDial(t *testing.T) {
	var originHits atomic.Int32
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHits.Add(1)
		_, _ = fmt.Fprint(w, "via-socks")
	}))
	defer origin.Close()

	var socksConns atomic.Int32
	socksLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer socksLn.Close()

	// Minimal SOCKS5 no-auth server (CONNECT only).
	go serveTestSOCKS5(t, socksLn, &socksConns)

	proxyURL := "socks5://" + socksLn.Addr().String()
	client, err := HTTPClient(Config{URL: proxyURL}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Get(origin.URL + "/x")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "via-socks" {
		t.Fatalf("body %q", body)
	}
	if socksConns.Load() < 1 {
		t.Fatal("SOCKS proxy saw no CONNECT")
	}
	if originHits.Load() < 1 {
		t.Fatal("origin not reached")
	}
}

func serveTestSOCKS5(t *testing.T, ln net.Listener, hits *atomic.Int32) {
	t.Helper()
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			defer conn.Close()
			if err := handleSOCKS5(conn, hits); err != nil {
				return
			}
		}(c)
	}
}

func handleSOCKS5(conn net.Conn, hits *atomic.Int32) error {
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 512)

	// greeting
	n, err := io.ReadAtLeast(conn, buf, 2)
	if err != nil {
		return err
	}
	if buf[0] != 0x05 {
		return fmt.Errorf("ver")
	}
	nmethods := int(buf[1])
	need := 2 + nmethods
	for n < need {
		m, err := conn.Read(buf[n:need])
		if err != nil {
			return err
		}
		n += m
	}
	// no auth
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	// request
	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return err
	}
	if buf[0] != 0x05 || buf[1] != 0x01 { // CONNECT
		return fmt.Errorf("cmd")
	}
	var host string
	var port uint16
	switch buf[3] {
	case 0x01: // IPv4
		if _, err := io.ReadFull(conn, buf[:4]); err != nil {
			return err
		}
		host = net.IP(buf[:4]).String()
	case 0x03: // domain
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return err
		}
		l := int(buf[0])
		if _, err := io.ReadFull(conn, buf[:l]); err != nil {
			return err
		}
		host = string(buf[:l])
	case 0x04: // IPv6
		if _, err := io.ReadFull(conn, buf[:16]); err != nil {
			return err
		}
		host = net.IP(buf[:16]).String()
	default:
		return fmt.Errorf("atyp")
	}
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return err
	}
	port = uint16(buf[0])<<8 | uint16(buf[1])

	target := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	remote, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		_, _ = conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return err
	}
	defer remote.Close()

	// success + bind addr placeholder
	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	// Count successful CONNECT before tunnel (avoids race with client close).
	hits.Add(1)

	// transparent tunnel
	errc := make(chan error, 2)
	go func() { _, e := io.Copy(remote, conn); errc <- e }()
	go func() { _, e := io.Copy(conn, remote); errc <- e }()
	<-errc
	return nil
}
