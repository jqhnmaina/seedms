package token

import "github.com/dgrijalva/jwt-go"

type Token struct {
	UserID int64
	Groups []string
	jwt.StandardClaims
}
