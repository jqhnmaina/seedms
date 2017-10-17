package testing

import (
	"database/sql"

	apiG "github.com/tomogoma/go-api-guard"
	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/api"
)

type DBMock struct {
	errors.NotFoundErrCheck

	ExpInsAPIKErr     error
	ExpAPIKsBUsrID    *api.Key
	ExpAPIKsBUsrIDErr error

	isInTx bool
}

func (db *DBMock) ExecuteTx(fn func(*sql.Tx) error) error {
	db.isInTx = true
	defer func() {
		db.isInTx = false
	}()
	return fn(new(sql.Tx))
}

func (db *DBMock) APIKeyByUserIDVal(userID string, key []byte) (apiG.Key, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpAPIKsBUsrID == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpAPIKsBUsrID, db.ExpAPIKsBUsrIDErr
}

func (db *DBMock) InsertAPIKey(userID string, key []byte) (apiG.Key, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsAPIKErr != nil {
		return nil, db.ExpInsAPIKErr
	}
	return &api.Key{ID: currentID(), UserID: userID, Val: key}, db.ExpInsAPIKErr
}
