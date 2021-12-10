package security

import (
	"github.com/mytja/SMTP2/helpers"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pass string) (string, error) {
	passbyte := helpers.StringToBytearray(pass)
	password, err := bcrypt.GenerateFromPassword(passbyte, 14)
	if err != nil {
		return "", err
	}
	passstr := helpers.BytearrayToString(password)
	return passstr, nil
}

func CheckHash(pass string, hashedPass string) bool {
	err := bcrypt.CompareHashAndPassword(helpers.StringToBytearray(hashedPass), helpers.StringToBytearray(pass))
	return err == nil
}
