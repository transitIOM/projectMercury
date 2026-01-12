package middleware

import (
	"crypto"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var ExpectedHash string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}
	s, exists := os.LookupEnv("API_KEY_HASH")
	if exists && s != "" {
		ExpectedHash = s
		return
	}
	s, exists = os.LookupEnv("API_KEY")
	if exists && s != "" {
		h := crypto.SHA256.New()
		h.Write([]byte(s))
		ExpectedHash = hex.EncodeToString(h.Sum(nil))
		log.Warn("pre-hashed API key preferred; please update ENV configuration")
		return
	}
	err := errors.New("failed to get API key")
	log.Fatal(err)
}

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Authenticating request")
		headerKey := r.Header.Get("X-API-Key")
		if headerKey == "" {
			log.Debug("X-API-Key header missing")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		h := crypto.SHA256.New()
		h.Write([]byte(headerKey))

		userHash := hex.EncodeToString(h.Sum(nil))

		if subtle.ConstantTimeCompare([]byte(userHash), []byte(ExpectedHash)) != 1 {
			log.Debug("API key hash mismatch")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		log.Debug("Authentication successful")
		// add info to context
		//ctx := context.WithValue(r.Context(), apiKeyCtxKey, "admin")

		next.ServeHTTP(w, r)
	})
}
