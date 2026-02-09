package listings

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"app.root/db"
	"app.root/httpjson"
)

type CountHandler struct {
	DB *sql.DB
}

func (h *CountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	q := db.New(h.DB)

	n, err := q.CountVisibleListings(ctx)
	if err != nil {
		http.Error(w, "count failed", http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]int64{
		"count": n,
	})
}
