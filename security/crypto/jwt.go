package crypto

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/mytja/SMTP2/helpers/constants"
	"net/http"
	"time"
)

func GetJWTFromUserPass(email string, pass string) (string, error) {
	// IMPORTANT: Password MUST BE hashed

	expirationTime := time.Now().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"iss":   constants.JwtIssuer,
		"exp":   expirationTime.Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(constants.JwtSigningKey)

	return tokenString, err
}

func CheckJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return constants.JwtSigningKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] == constants.JwtIssuer {
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
	email := fmt.Sprint(j["email"])
	return true, email, nil
}
