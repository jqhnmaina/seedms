package http

import (
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/jwt"
)

type JWTHelper interface {
	errors.ToHTTPResponser

	Valid(token string) (*jwt.Claim, error)
}
