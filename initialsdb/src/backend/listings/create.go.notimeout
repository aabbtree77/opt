package listings

import (
	"crypto/sha256"
	"database/sql"
	"net/http"
	"strings"

	"app.root/config"
	"app.root/db"
	"app.root/guards"
	"app.root/httpjson"
)

type CreateHandler struct {
	DB     *sql.DB
	Cfg    *config.Config
	Guards []guards.Guard
}

type createListingRequest struct {
	Text string `json:"text"`
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.WriteError(w, http.StatusMethodNotAllowed, "INVALID_INPUT", "method not allowed")
		return
	}

	for _, g := range h.Guards {
		if !g.Check(r) {
			httpjson.Forbidden(w, "RATE_LIMITED", "request blocked")
			return
		}
	}

	var req createListingRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.BadRequest(w, "INVALID_INPUT", "invalid json body")
		return
	}

	body := strings.TrimSpace(req.Text)
	if body == "" {
		httpjson.BadRequest(w, "INVALID_INPUT", "empty body")
		return
	}

	ip := strings.TrimSpace(guards.GetIP(r))
	ipHash := sha256.Sum256([]byte(ip + h.Cfg.ServerSalt))

	store := db.NewStore(h.DB)

	//Very bad idea to IP rate limit based on DB, it actually facilitates DDOS!
	//Replaced this with in RAM IP rate limiter, chatgpt suggests using Redis later on.
	/*
		count, err := store.CountRecentListingsByIP(r.Context(), ipHash[:])
		if err != nil {
			httpjson.InternalError(w, "db error")
			return
		}
		if count >= 1000 {
			httpjson.TooManyRequests(w, "RATE_LIMITED", "too many posts")
			return
		}
	*/

	listing, err := store.CreateListing(r.Context(), db.CreateListingParams{
		Body:   body,
		IpHash: ipHash[:],
	})
	if err != nil {
		httpjson.InternalError(w, "db error")
		return
	}

	httpjson.WriteCreated(w, listing)
}
