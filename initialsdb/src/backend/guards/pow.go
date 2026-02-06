package guards

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

/*
────────────────────────────────────────────────────────────
Config
────────────────────────────────────────────────────────────
*/

type PowConfig struct {
	Enable     bool
	Difficulty uint8
	TTL        time.Duration
	SecretKey  []byte
}

/*
────────────────────────────────────────────────────────────
Challenge handler
────────────────────────────────────────────────────────────
*/

type PoWHandler struct {
	Cfg PowConfig
	Key []byte
}

func NewPoWHandler(cfg PowConfig) *PoWHandler {
	return &PoWHandler{
		Cfg: cfg,
		Key: cfg.SecretKey,
	}
}

type challengePayload struct {
	Challenge  string `json:"challenge"`
	Difficulty uint8  `json:"difficulty"`
	TTLSecs    int64  `json:"ttl_secs"`
	Token      string `json:"token"`
}

func (h *PoWHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.Cfg.Enable {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// never cache PoW challenges
	w.Header().Set("Cache-Control", "no-store")

	ch := make([]byte, 16)
	_, _ = rand.Read(ch)
	chStr := base64.RawStdEncoding.EncodeToString(ch)

	now := time.Now().Unix()
	exp := now + int64(h.Cfg.TTL.Seconds())

	expBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expBytes, uint64(exp))

	mac := hmac.New(sha256.New, h.Key)
	mac.Write([]byte(chStr))
	mac.Write(expBytes)
	hmacPart := mac.Sum(nil)

	token := base64.RawStdEncoding.EncodeToString(hmacPart) +
		"." +
		base64.RawStdEncoding.EncodeToString(expBytes)

	resp := challengePayload{
		Challenge:  chStr,
		Difficulty: h.Cfg.Difficulty,
		TTLSecs:    exp - now,
		Token:      token,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

/*
────────────────────────────────────────────────────────────
Guard
────────────────────────────────────────────────────────────
*/

type PoWGuard struct {
	Cfg PowConfig
	Key []byte
}

func NewPoWGuard(cfg PowConfig) *PoWGuard {
	return &PoWGuard{
		Cfg: cfg,
		Key: cfg.SecretKey,
	}
}

func (g *PoWGuard) Check(r *http.Request) bool {
	if !g.Cfg.Enable {
		return true
	}

	challenge := r.Header.Get("X-PoW-Challenge")
	nonce := r.Header.Get("X-PoW-Nonce")
	token := r.Header.Get("X-PoW-Token")

	if challenge == "" || nonce == "" || token == "" {
		return false
	}

	// hard cap nonce size
	if len(nonce) > 64 {
		return false
	}

	return VerifyPoW(g.Cfg, g.Key, challenge, nonce, token) == nil
}

/*
────────────────────────────────────────────────────────────
Verification
────────────────────────────────────────────────────────────
*/

func VerifyPoW(cfg PowConfig, key []byte, challenge, nonce, token string) error {
	if !cfg.Enable {
		return nil
	}

	if cfg.Difficulty == 0 || cfg.TTL <= 0 || len(key) == 0 {
		return errors.New("invalid pow config")
	}

	exp, err := parseToken(key, challenge, token)
	if err != nil {
		return err
	}

	if time.Now().Unix() > exp {
		return errors.New("challenge expired")
	}

	if !checkDifficulty(challenge, nonce, cfg.Difficulty) {
		return errors.New("invalid pow")
	}

	return nil
}

func parseToken(key []byte, challenge, token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, errors.New("invalid token format")
	}

	hmacPart, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, errors.New("bad hmac encoding")
	}

	expRaw, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil || len(expRaw) != 8 {
		return 0, errors.New("bad exp encoding")
	}

	exp := int64(binary.BigEndian.Uint64(expRaw))

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(challenge))
	mac.Write(expRaw)
	expected := mac.Sum(nil)

	if !hmac.Equal(expected, hmacPart) {
		return 0, errors.New("bad hmac")
	}

	return exp, nil
}

func checkDifficulty(challenge, nonce string, difficulty uint8) bool {
	chBytes, err := base64.RawStdEncoding.DecodeString(challenge)
	if err != nil {
		return false
	}

	sum := sha256.Sum256(append(chBytes, nonce...))

	var zeros uint8
	for _, b := range sum {
		for bit := 7; bit >= 0; bit-- {
			if (b>>bit)&1 == 0 {
				zeros++
				if zeros >= difficulty {
					return true
				}
			} else {
				return false
			}
		}
	}
	return false
}
