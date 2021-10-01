package crypto

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/objects"
	"github.com/mytja/SMTP2/sql"
	"net/http"
)

func GetJWTFromUserPass(email string, pass string) (string, error) {
	// IMPORTANT: Password MUST BE hashed

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"pass":  pass,
		"iss":   constants.JWT_ISSUER,
		//"exp": 3 * 24 * 60 * 60,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(constants.JWT_SIGNING_KEY)

	return tokenString, err
}

func CheckJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return constants.JWT_SIGNING_KEY, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] == constants.JWT_ISSUER {
			return claims, nil
		}
		return nil, errors.New("JWT issuer isn't correct")
	} else {
		return nil, err
	}
}

func CheckUser(r *http.Request) (bool, string, error) {
	token := r.Header.Get("X-Login-Token")
	if token == "" {
		return false, "", errors.New(constants.ERR_NOJWTPROVIDED)
	}
	j, err := CheckJWT(token)
	if err != nil {
		return false, "", err
	}
	email := j["email"]
	pass := j["pass"]
	var user objects.User
	err = sql.DB.GetDB().Get(&user, "SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		return false, "", err
	}
	if pass == user.Password {
		return true, user.Email, nil
	}
	return false, "", nil
}
