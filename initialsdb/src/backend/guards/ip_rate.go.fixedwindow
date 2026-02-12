package guards

import (
	"net"
	"net/http"
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
// Guard
// ──────────────────────────────────────────────
//

// IPRateGuard is a fixed-window per-IP rate limiting guard.
// It is silent: it does not write responses or redirect.
type IPRateGuard struct {
	mu              sync.Mutex
	entries         map[string]int
	maxRequests     int
	window          time.Duration
	currWindowStart time.Time
	enable          bool
}

func NewIPRateGuard(cfg IPRateLimiterConfig) *IPRateGuard {
	return &IPRateGuard{
		enable:          cfg.Enable,
		maxRequests:     cfg.MaxRequests,
		window:          cfg.Window,
		entries:         make(map[string]int),
		currWindowStart: time.Now(),
	}
}

func (g *IPRateGuard) Check(r *http.Request) bool {
	if !g.enable {
		return true
	}

	ip := GetIP(r)
	if ip == "" || g.maxRequests <= 0 || g.window <= 0 {
		return true
	}

	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	// Window rollover
	if now.Sub(g.currWindowStart) >= g.window {
		g.entries = make(map[string]int)
		g.currWindowStart = now
	}

	if g.entries[ip] >= g.maxRequests {
		return false
	}

	g.entries[ip]++
	return true
}

//
// ──────────────────────────────────────────────
// IP extraction
// ──────────────────────────────────────────────
//

func GetIP(r *http.Request) string {
	// DEV / TEST override
	if ip := r.Header.Get("X-Test-IP"); ip != "" {
		return ip
	}

	// Behind proxy (assumed trusted at infra level)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return ""
}
