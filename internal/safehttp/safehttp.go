//go:build !js || !wasm

// Package safehttp provides outbound HTTP requests that cannot reach private
// network addresses.
package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

const requestTimeout = 30 * time.Second

var blockedNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
	netip.MustParsePrefix("2001:db8::/32"),
}

// NewClient returns an HTTP client restricted to public network addresses.
func NewClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = publicDialContext(dialer, net.DefaultResolver)

	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after %d redirects", len(via))
			}
			return ValidateURL(req.URL)
		},
	}
}

// NewRequest creates an HTTP request after checking its URL policy.
func NewRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if err := ValidateURL(parsed); err != nil {
		return nil, err
	}
	return http.NewRequestWithContext(ctx, method, parsed.String(), nil)
}

// ValidateURL accepts absolute HTTP and HTTPS URLs without embedded
// credentials.
func ValidateURL(parsed *url.URL) error {
	if parsed == nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" {
		return fmt.Errorf("URL must be absolute")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
	}
	if parsed.User != nil {
		return fmt.Errorf("URL credentials are not allowed")
	}
	return nil
}

func publicDialContext(dialer *net.Dialer, resolver *net.Resolver) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid network address %q: %w", address, err)
		}

		resolved, err := resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", host, err)
		}
		if len(resolved) == 0 {
			return nil, fmt.Errorf("resolve %q: no addresses", host)
		}

		var lastErr error
		for _, addressIP := range resolved {
			if !isPublic(addressIP) {
				return nil, fmt.Errorf("network address %q is not public", addressIP)
			}
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(addressIP.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		return nil, fmt.Errorf("connect to %q: %w", host, lastErr)
	}
}

func isPublic(address netip.Addr) bool {
	if !address.IsValid() {
		return false
	}
	address = address.Unmap()
	for _, network := range blockedNetworks {
		if network.Contains(address) {
			return false
		}
	}
	return address.IsGlobalUnicast()
}
