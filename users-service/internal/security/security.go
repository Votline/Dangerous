// Package security provides security functions
// hashing and compare passwords
package security

import (
	"unsafe"

	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
	passwdBytes := unsafe.Slice(unsafe.StringData(password), len(password))
	hash, err := bcrypt.GenerateFromPassword(passwdBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Check(password, hash string) error {
	passwdBytes := unsafe.Slice(unsafe.StringData(password), len(password))
	hashBytes := unsafe.Slice(unsafe.StringData(hash), len(hash))
	err := bcrypt.CompareHashAndPassword(hashBytes, passwdBytes)
	if err != nil {
		return err
	}
	return nil
}
