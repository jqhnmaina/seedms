package jwt

import (
	"github.com/tomogoma/go-typed-errors"
)

type Helper struct {
	errors.ErrToHTTP

	validater Validator
}

func NewHelper(v Validator) (*Helper, error) {

	if v == nil {
		return nil, errors.New("nil Validator")
	}

	return &Helper{validater: v}, nil
}

func (h Helper) Valid(token string) (*Claim, error) {

	claim := &Claim{}

	if _, err := h.validater.Validate(token, claim); err != nil {
		if h.validater.IsAuthError(err) {
			return nil, errors.NewAuth(err)
		}
		if h.validater.IsForbiddenError(err) {
			return nil, errors.NewForbidden(err)
		}
		if h.validater.IsUnauthorizedError(err) {
			return nil, errors.NewUnauthorized(err)
		}
		return nil, errors.Newf("Validate token: %v", err)
	}

	return claim, nil
}
