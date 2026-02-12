package guards

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//
// ──────────────────────────────────────────────
// Config
// ──────────────────────────────────────────────
//

type IPRateLimiterConfig struct {
	Enable      bool
	MaxRequests int
	Window      time.Duration
}

//
// ──────────────────────────────────────────────
// Guard (sliding window per IP)
// ──────────────────────────────────────────────
//

type ipEntry struct {
	count     int
	windowEnd time.Time
}

type IPRateGuard struct {
	mu          sync.Mutex
	entries     map[string]*ipEntry
	maxRequests int
	window      time.Duration
	enable      bool
}

func NewIPRateGuard(cfg IPRateLimiterConfig) *IPRateGuard {
	return &IPRateGuard{
		enable:      cfg.Enable,
		maxRequests: cfg.MaxRequests,
		window:      cfg.Window,
		entries:     make(map[string]*ipEntry),
	}
}

func (g *IPRateGuard) Check(r *http.Request) bool {
	if !g.enable {
		return true
	}

	if g.maxRequests <= 0 || g.window <= 0 {
		return true
	}

	ip := normalizeIP(GetIP(r))
	if ip == "" {
		return true
	}

	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	entry, exists := g.entries[ip]

	if !exists || now.After(entry.windowEnd) {
		// start new window for this IP
		g.entries[ip] = &ipEntry{
			count:     1,
			windowEnd: now.Add(g.window),
		}
		g.cleanup(now)
		return true
	}

	if entry.count >= g.maxRequests {
		return false
	}

	entry.count++
	return true
}

//
// ──────────────────────────────────────────────
// Cleanup (lazy, cheap)
// ──────────────────────────────────────────────
//

func (g *IPRateGuard) cleanup(now time.Time) {
	// Lazy cleanup: remove expired IP buckets
	for ip, entry := range g.entries {
		if now.After(entry.windowEnd) {
			delete(g.entries, ip)
		}
	}
}

//
// ──────────────────────────────────────────────
// IP extraction (Caddy-safe)
// ──────────────────────────────────────────────
//

/*
func GetIP(r *http.Request) string {
	// DEV / TEST override
	if ip := strings.TrimSpace(r.Header.Get("X-Test-IP")); ip != "" {
		return ip
	}

	// Properly parse X-Forwarded-For
	// It may contain multiple IPs: client, proxy1, proxy2
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return ""
}
*/

func GetIP(r *http.Request) string {
	// DEV / TEST override (validated)
	if ip := strings.TrimSpace(r.Header.Get("X-Test-IP")); ip != "" {
		if parsed := net.ParseIP(ip); parsed != nil {
			return parsed.String()
		}
	}

	// Behind trusted reverse proxy (Caddy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		candidate := strings.TrimSpace(parts[0]) // first = original client

		if parsed := net.ParseIP(candidate); parsed != nil {
			return parsed.String()
		}
	}

	// Direct connection fallback
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if parsed := net.ParseIP(host); parsed != nil {
			return parsed.String()
		}
	}

	return ""
}

//
// ──────────────────────────────────────────────
// IP normalization (IPv6-safe)
// ──────────────────────────────────────────────
//

func normalizeIP(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ""
	}
	return parsed.String()
}
