package middlewares

import (
	"os"
	"time"
	"user/internal/models"

	"github.com/dgrijalva/jwt-go"
)

func CreateToken(user models.User) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenExpirationString:= os.Getenv("TOKEN_EXPIRATION")
	
	tokenExpiration, err := time.ParseDuration(tokenExpirationString)
	if err != nil {
		return "", err
	}
	
	claims := jwt.MapClaims{
		"userId":         	user.UserId,
		"expirationDate":  	time.Now().Add(tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
