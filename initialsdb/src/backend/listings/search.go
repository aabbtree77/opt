package listings

import (
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app.root/db"
	"app.root/guards"
	"app.root/httpjson"
)

type SearchHandler struct {
	DB     *sql.DB
	Guards []guards.Guard
}

type listingResult struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type searchResponse struct {
	Items      []listingResult `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	limit := int32(30)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}

	cursor := r.URL.Query().Get("cursor")
	store := db.NewStore(h.DB)

	var rows []listingResult

	if cursor == "" {
		res, err := store.SearchListingsFirstPage(
			r.Context(),
			db.SearchListingsFirstPageParams{
				Column1: q,
				Limit:   limit,
			},
		)
		if err != nil {
			httpjson.InternalError(w, "db error")
			return
		}

		rows = make([]listingResult, 0, len(res))
		for _, r := range res {
			rows = append(rows, listingResult{
				ID:        r.ID,
				Body:      r.Body,
				CreatedAt: r.CreatedAt,
			})
		}
	} else {
		createdAt, id, ok := decodeCursor(cursor)
		if !ok {
			httpjson.BadRequest(w, "INVALID_INPUT", "invalid cursor")
			return
		}

		res, err := store.SearchListingsAfterCursor(
			r.Context(),
			db.SearchListingsAfterCursorParams{
				Column1:   q,
				CreatedAt: createdAt,
				ID:        id,
				Limit:     limit,
			},
		)
		if err != nil {
			httpjson.InternalError(w, "db error")
			return
		}

		rows = make([]listingResult, 0, len(res))
		for _, r := range res {
			rows = append(rows, listingResult{
				ID:        r.ID,
				Body:      r.Body,
				CreatedAt: r.CreatedAt,
			})
		}
	}

	resp := searchResponse{
		Items: rows,
	}

	if len(rows) == int(limit) {
		last := rows[len(rows)-1]
		resp.NextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	httpjson.WriteOK(w, resp)
}

func encodeCursor(t time.Time, id int64) string {
	payload := strconv.FormatInt(t.UnixNano(), 10) + ":" + strconv.FormatInt(id, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeCursor(s string) (time.Time, int64, bool) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, 0, false
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return time.Time{}, 0, false
	}

	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, 0, false
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, false
	}

	return time.Unix(0, ns).UTC(), id, true
}
