package jwt

import (
	"errors"
	"flag"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

var (
	secretKeyFlag = flag.String("secretKey", "", "secret key")
)

func GenerateJWT(email string, id int) (string, error) {
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

func JWTMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		authHeader := string(ctx.Request.Header.Peek("Authorization"))
		if authHeader == "" {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBody([]byte(`{"error": "missing token"}`))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBody([]byte(`{"error": "invalid token format"}`))
			return
		}

		tokenStr := parts[1]

		claims, err := parseJWT(tokenStr)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBody([]byte(`{"error": "` + err.Error() + `"}`))
			return
		}

		ctx.SetUserValue("userID", claims["userID"])
		ctx.SetUserValue("email", claims["email"])

		next(ctx)
	}
}

func parseJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(*secretKeyFlag), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	expiration, ok := claims["exp"].(float64)
	if !ok || int64(expiration) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
