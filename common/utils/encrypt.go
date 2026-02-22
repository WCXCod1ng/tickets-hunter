package utils

import "golang.org/x/crypto/bcrypt"

// Encrypt takes a plain text string and returns its bcrypt hash.
func Encrypt(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(plain),
		bcrypt.DefaultCost, // 一般够用，默认是 10
	)
	return string(bytes), err
}

// Verify compares the plain string with the encrypted hash and returns true if they match.
func Verify(plain, encrypted string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(encrypted),
		[]byte(plain),
	)
	return err == nil
}
