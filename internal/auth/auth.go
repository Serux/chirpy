package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	ret, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(ret), nil
}

func CheckPasswordHash(password, hash string) error {
	ret, _ := HashPassword(password)
	return bcrypt.CompareHashAndPassword([]byte(ret), []byte(hash))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   userID.String()})
	signtok, _ := token.SignedString([]byte(tokenSecret))

	return signtok, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	jwtfunc := func(t *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil }
	token, err := jwt.ParseWithClaims(tokenString, &claims, jwtfunc)
	if err != nil {
		return uuid.UUID{}, err
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	if t, err := token.Claims.GetExpirationTime(); time.Now().After(t.Time) {
		return uuid.UUID{}, err
	}

	uid, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return uid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("NO AUTH")
	}
	authFields := strings.Fields(auth)
	if len(authFields) != 2 && authFields[0] != "Bearer" {
		return "", fmt.Errorf("WRONG AUTH")
	}
	return authFields[1], nil

}

func MakeRefreshToken() (string, error) {
	sli := make([]byte, 32)
	_, err := rand.Read(sli)
	if err != nil {
		return "", err
	}
	str := hex.EncodeToString(sli)

	return str, nil

}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("NO AUTH")
	}
	authFields := strings.Fields(auth)
	if len(authFields) != 2 && authFields[0] != "ApiKey" {
		return "", fmt.Errorf("WRONG AUTH")
	}
	return authFields[1], nil
}
