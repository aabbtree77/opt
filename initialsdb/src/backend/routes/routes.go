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
	// Common guards (reads + writes)
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

	// Body size only meaningful for endpoints with request body.
	var bodyGuard []guards.Guard
	if cfg.BodySizeLimiter.Enable {
		bodyGuard = append(bodyGuard,
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
	// Listings: search (GET)
	// ────────────────────────────────────────

	mux.Handle("/api/listings/search",
		&listings.SearchHandler{
			DB:     db,
			Guards: guardsCommon,
		},
	)

	// ────────────────────────────────────────
	// Listings: create (POST)
	// ────────────────────────────────────────

	guardsCreate := append([]guards.Guard{}, guardsCommon...)
	guardsCreate = append(guardsCreate, bodyGuard...)

	if cfg.ProofOfWork.Enable {
		guardsCreate = append(guardsCreate, guards.NewPoWGuard(powCfg))
		mux.Handle("/pow/challenge", guards.NewPoWHandler(powCfg))
	}

	mux.Handle("/api/listings/create",
		&listings.CreateHandler{
			DB:     db,
			Cfg:    cfg,
			Guards: guardsCreate,
		},
	)

	// ────────────────────────────────────────
	// Listings: count (GET)
	// ────────────────────────────────────────

	mux.Handle("/api/listings/count",
		&listings.CountHandler{
			DB:     db,
			Guards: guardsCommon,
		},
	)

	// ────────────────────────────────────────
	// SPA fallback
	// ────────────────────────────────────────

	mux.Handle("/", spa.SPAHandler{
		Dir: "web",
	})
}
