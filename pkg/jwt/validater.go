package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/go-typed-errors"
)

type Validator interface {
	errors.IsAuthErrChecker

	Validate(token string, cs jwt.Claims) (*jwt.Token, error)
}
