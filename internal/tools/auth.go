package tools

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func ValidateCredentials(token string) bool {
	truth := os.Getenv("TOKEN")
	if token == truth {
		return true
	}
	return false
}

func ConstructJWT(aud []string, exp time.Time, iat time.Time, iss string) (string, error) {
	nexp := jwt.NewNumericDate(exp)
	niat := jwt.NewNumericDate(iat)
	claims := jwt.RegisteredClaims{
		Audience:  aud,
		ExpiresAt: nexp,
		IssuedAt:  niat,
		Issuer:    iss,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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
