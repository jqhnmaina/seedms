package testing

import (
	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/go-typed-errors"
)

type GuardMock struct {
	errors.AuthErrCheck

	ExpAPIKValidUsrID string
	ExpAPIKValidErr   error
	ExpNewAPIK        *api.Key
	ExpNewAPIKErr     error
}

func (g *GuardMock) APIKeyValid(key []byte) (string, error) {
	return g.ExpAPIKValidUsrID, g.ExpAPIKValidErr
}
func (g *GuardMock) NewAPIKey(userID string) (*api.Key, error) {
	return g.ExpNewAPIK, g.ExpNewAPIKErr
}
