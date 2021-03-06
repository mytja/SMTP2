package security

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/sql"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type securityImpl struct {
	db     sql.SQL
	logger *zap.SugaredLogger
}

type Security interface {
	CheckUser(r *http.Request) (bool, string, error)
	GetProtocolFromDomain(todomain string) (string, error)

	VerifyEmailServer(mail sql.ReceivedMessage) error
	VerifyEmailSender(mail sql.ReceivedMessage) error
	VerifyMessage(mail sql.ReceivedMessage) (bool, error)
}

func NewSecurity(db sql.SQL, logger *zap.SugaredLogger) Security {
	return &securityImpl{db: db, logger: logger}
}

func GetJWTFromUserPass(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"iss":   helpers.JwtIssuer,
		"exp":   expirationTime.Unix(),
	})

	return token.SignedString(helpers.JwtSigningKey)
}

func CheckJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return helpers.JwtSigningKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] == helpers.JwtIssuer {
			return claims, nil
		}
		return nil, errors.New("JWT issuer isn't correct")
	} else {
		return nil, err
	}
}

func (security *securityImpl) CheckUser(r *http.Request) (bool, string, error) {
	token := r.Header.Get("X-Login-Token")
	if token == "" {
		return false, "", errors.New("unauthenticated")
	}
	j, err := CheckJWT(token)
	if err != nil {
		return false, "", err
	}
	email := fmt.Sprint(j["email"])
	user, err := security.db.GetUserByEmail(email)
	if err != nil {
		return false, "", err
	}
	if email != user.Email {
		return false, "", errors.New("no user with this email was located in database")
	}
	return true, email, nil
}
