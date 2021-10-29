package reader

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/guregu/null"
	"log"
	"time"
)

func NewStandardClaims(expiresAt int64) jwt.StandardClaims {
	return jwt.StandardClaims{
		ExpiresAt: expiresAt,
		IssuedAt:  time.Now().Unix(),
		Issuer:    "cn.ftacademy.reader",
	}
}

type PassportClaims struct {
	FtcID   string      `json:"fid"`
	UnionID null.String `json:"wid"`
	jwt.StandardClaims
}

type Passport struct {
	Account
	ExpiresAt int64  `json:"expiresAt"`
	Token     string `json:"token"`
}

func NewPassport(a Account, signingKey []byte) (Passport, error) {
	claims := PassportClaims{
		FtcID:          a.FtcID,
		UnionID:        a.UnionID,
		StandardClaims: NewStandardClaims(time.Now().Unix() + 86400*7),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)

	if err != nil {
		return Passport{}, err
	}

	return Passport{
		Account:   a,
		ExpiresAt: claims.ExpiresAt,
		Token:     ss,
	}, nil
}

func ParsePassportClaims(ss string, key []byte) (PassportClaims, error) {
	token, err := jwt.ParseWithClaims(
		ss,
		&PassportClaims{},
		func(token *jwt.Token) (i interface{}, err error) {
			return key, nil
		})

	if err != nil {
		log.Printf("Parsing JWT error: %v", err)
		return PassportClaims{}, err
	}

	log.Printf("Claims: %v", token.Claims)

	// NOTE: token.Claims is an interface, so it is a pointer, not a value type.
	if claims, ok := token.Claims.(*PassportClaims); ok {
		return *claims, nil
	}
	return PassportClaims{}, errors.New("wrong JWT claims")
}