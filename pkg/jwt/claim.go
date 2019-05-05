package jwt

import "github.com/dgrijalva/jwt-go"

type Claim struct {
	UsrID string
	Group Group
	jwt.StandardClaims
}
