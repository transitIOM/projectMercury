package tools

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

func ValidateCredentials(token string) bool {
	return token == os.Getenv("TOKEN")
}

func ConstructJWT(claims map[string]interface{}) (string, error) {
	tokenClaims := jwt.MapClaims{}
	for key, value := range claims {
		tokenClaims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	jwtKey := []byte(os.Getenv("JWT_KEY"))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseJWT(tokenString string) (*jwt.Token, error) {
	jwtKey := []byte(os.Getenv("JWT_KEY"))

	return jwt.ParseWithClaims(tokenString, jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		},
	)
}

func IsAdminToken(token *jwt.Token) bool {
	return token.Claims.(jwt.MapClaims)["aud"] == "admin"
}
