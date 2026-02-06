package guards

import "net/http"

//
// ──────────────────────────────────────────────
// Guard
// ──────────────────────────────────────────────
//

// BodySizeGuard limits the maximum request body size.
// It is silent: it does not write responses or redirect.
type BodySizeGuard struct {
	enable  bool
	maxSize int64
}

func NewBodySizeGuard(enable bool, maxBytes int64) *BodySizeGuard {
	return &BodySizeGuard{
		enable:  enable,
		maxSize: maxBytes,
	}
}

func (g *BodySizeGuard) Check(r *http.Request) bool {
	if !g.enable || g.maxSize <= 0 {
		return true
	}

	// Fast path: known and already too large
	if r.ContentLength >= 0 && r.ContentLength > g.maxSize {
		return false
	}

	// Unknown size (e.g. chunked):
	// allow and let handlers fail naturally if parsing exceeds limits.
	return true
}
