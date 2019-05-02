package http

import (
	"bitbucket.org/rfhkenya/qms-server/pkg/jwt"
	"github.com/tomogoma/go-typed-errors"
)

type JWTHelper interface {
	errors.ToHTTPResponser

	Valid(token string) (*jwt.Claim, error)
}
