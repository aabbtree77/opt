-- =====================================================
-- LISTINGS
-- =====================================================

-- name: CreateListing :one
INSERT INTO listings (
    body,
    ip_hash
) VALUES (
    $1,
    $2
)
RETURNING
    id,
    body,
    is_hidden,
    created_at;


-- =====================================================
-- LISTINGS SEARCH (KEYSET PAGINATION)
-- =====================================================

-- name: SearchListingsFirstPage :many
SELECT
    id,
    body,
    created_at
FROM listings
WHERE
    is_hidden = FALSE
    AND (
        $1::text IS NULL
        OR body_tsv @@ plainto_tsquery('simple', $1)
    )
ORDER BY created_at DESC, id DESC
LIMIT $2;


-- name: SearchListingsAfterCursor :many
SELECT
    id,
    body,
    created_at
FROM listings
WHERE
    is_hidden = FALSE
    AND (
        $1::text IS NULL
        OR body_tsv @@ plainto_tsquery('simple', $1)
    )
    AND (
        created_at < $2
        OR (created_at = $2 AND id < $3)
    )
ORDER BY created_at DESC, id DESC
LIMIT $4;


-- =====================================================
-- BOT / RATE LIMITING HELPERS
-- =====================================================

-- name: CountRecentListingsByIP :one
SELECT COUNT(*)
FROM listings
WHERE
    ip_hash = $1
    AND created_at >= now() - INTERVAL '1 hour';


-- name: TouchListingsByIP :exec
UPDATE listings
SET ip_hash = ip_hash
WHERE ip_hash = $1;


-- =====================================================
-- Global counter for DB entries above search
-- =====================================================

-- name: CountVisibleListings :one
SELECT COUNT(*)::bigint
FROM listings
WHERE is_hidden = FALSE;
