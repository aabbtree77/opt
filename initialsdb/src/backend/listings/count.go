package listings

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"app.root/db"
	"app.root/guards"
	"app.root/httpjson"
)

type CountHandler struct {
	DB     *sql.DB
	Guards []guards.Guard
}

func (h *CountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.WriteError(w, http.StatusMethodNotAllowed, "INVALID_INPUT", "method not allowed")
		return
	}

	for _, g := range h.Guards {
		if !g.Check(r) {
			httpjson.Forbidden(w, "RATE_LIMITED", "request blocked")
			return
		}
	}

	/*
		Add timeout to prevent request goroutine block indefinitely
		if DB stalls or network hiccup for some strange reason:

		query auto-cancels

		DB receives cancel signal

		goroutine freed

		client gets 500
	*/
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	q := db.New(h.DB)

	n, err := q.CountVisibleListings(ctx)
	if err != nil {
		httpjson.InternalError(w, "count failed")
		return
	}

	httpjson.WriteOK(w, map[string]int64{
		"count": n,
	})
}
