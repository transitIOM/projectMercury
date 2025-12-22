package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func GetAdminToken(w http.ResponseWriter, r *http.Request) {

	requestToken := r.Form.Get("token")

	if requestToken == "" {
		err := errors.New("token was either empty or not a string")
		log.Info(err)
		api.RequestErrorHandler(w, err)
		return
	}

	if requestToken != os.Getenv("TOKEN") {
		err := errors.New("unauthorized: incorrect token")
		api.UnauthorizedErrorHandler(w, err)
		return
	}

	claims := map[string]interface{}{
		"aud":     []string{"admin"},
		"exp":     jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		"iat":     jwt.NewNumericDate(time.Now()),
		"iss":     "github.com/transitiom/projectMercury",
		"user_id": "admin",
	}

	token, err := tools.ConstructJWT(claims)
	if err != nil {
		log.Warn(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetAuthTokenResponse{
		Code:  http.StatusOK,
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}
