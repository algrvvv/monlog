package user

import "golang.org/x/crypto/bcrypt"

func GenerateHashForPass(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(pass, hash string) bool {
	return nil == bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}
