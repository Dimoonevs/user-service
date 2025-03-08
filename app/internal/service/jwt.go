package service

import (
	"errors"
	"flag"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	secretKeyFlag = flag.String("secretKey", "", "secret key")
)

func generateJWT(email string, id int) (string, error) {
	claims := jwt.MapClaims{
		"userID": id,
		"email":  email,
		"exp":    time.Now().Add(time.Hour * 24 * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if secretKeyFlag == nil && *secretKeyFlag == "" {
		logrus.Errorf("secretKeyFlag is nil")
		return "", errors.New("secret key is required")
	}
	return token.SignedString([]byte(*secretKeyFlag))
}
