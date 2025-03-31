package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	pw := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckHashedPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
