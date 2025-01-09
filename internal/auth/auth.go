package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	ret, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(ret), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))
}
