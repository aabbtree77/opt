package routes

import (
	"database/sql"
	"net/http"

	"app.root/config"
	"app.root/guards"
	"app.root/listings"
	"app.root/spa"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *config.Config) {

	// ────────────────────────────────────────
	// Common guards
	// ────────────────────────────────────────

	var guardsCommon []guards.Guard

	if cfg.IPRateLimiter.Enable {
		guardsCommon = append(guardsCommon,
			guards.NewIPRateGuard(guards.IPRateLimiterConfig{
				Enable:      true,
				MaxRequests: cfg.IPRateLimiter.MaxRequests,
				Window:      cfg.IPRateLimiter.Window(),
			}),
		)
	}

	if cfg.BodySizeLimiter.Enable {
		guardsCommon = append(guardsCommon,
			guards.NewBodySizeGuard(true, cfg.BodySizeLimiter.MaxBytes),
		)
	}

	// ────────────────────────────────────────
	// Proof-of-Work (writes only)
	// ────────────────────────────────────────

	powCfg := guards.PowConfig{
		Enable:     cfg.ProofOfWork.Enable,
		Difficulty: cfg.ProofOfWork.Difficulty,
		TTL:        cfg.ProofOfWork.TTL(),
		SecretKey:  cfg.ProofOfWork.DecodedSecretKey,
	}

	// ────────────────────────────────────────
	// Listings: search
	// ────────────────────────────────────────

	mux.Handle("/api/listings/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		(&listings.SearchHandler{
			DB:     db,
			Guards: guardsCommon,
		}).ServeHTTP(w, r)
	}))

	// ────────────────────────────────────────
	// Listings: create
	// ────────────────────────────────────────

	var guardsCreate = append([]guards.Guard{}, guardsCommon...)

	if cfg.ProofOfWork.Enable {
		guardsCreate = append(guardsCreate, guards.NewPoWGuard(powCfg))
		mux.Handle("/pow/challenge", guards.NewPoWHandler(powCfg))
	}

	mux.Handle("/api/listings/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		(&listings.CreateHandler{
			DB:     db,
			Cfg:    cfg,
			Guards: guardsCreate,
		}).ServeHTTP(w, r)
	}))

	// ────────────────────────────────────────
	// SPA fallback
	// ────────────────────────────────────────

	mux.Handle("/", spa.SPAHandler{
		Dir: "web",
	})

	// ────────────────────────────────────────
	// Global counter of listings above search
	// ────────────────────────────────────────

	mux.Handle("/api/listings/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		(&listings.CountHandler{
			DB: db,
		}).ServeHTTP(w, r)
	}))

}
