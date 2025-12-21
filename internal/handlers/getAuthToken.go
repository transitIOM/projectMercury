package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/transitIOM/projectMercury/api"
	"github.com/transitIOM/projectMercury/internal/tools"
)

func GetAdminToken(w http.ResponseWriter, r *http.Request) {
	exp := time.Now().Add(7 * 24 * time.Hour)
	aud := []string{"admin"}
	iat := time.Now()
	iss := "github.com/transitiom/projectMercury"
	token, err := tools.ConstructJWT(aud, exp, iat, iss)
	if err != nil {
		fmt.Printf("err: %v", err)
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
