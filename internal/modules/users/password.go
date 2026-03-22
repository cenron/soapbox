package users

import "errors"

func HashPassword(plain string) (string, error) {
	return "", errors.New("not implemented")
}

func CheckPassword(hash, plain string) error {
	return errors.New("not implemented")
}
