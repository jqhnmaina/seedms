package model

import (
	"github.com/tomogoma/go-commons/database/cockroach"
	"database/sql"
	"fmt"
)

type Model struct {
	db *sql.DB
}

func New(dsnF cockroach.DSNFormatter) (*Model, error) {
	db, err := cockroach.DBConn(dsnF)
	if err != nil {
		return nil, fmt.Errorf("instantiating DB connection: %s", err)
	}
	// TODO SEEDMS replace seedsTable with own table definitions...
	err = cockroach.InstantiateDB(db, dsnF.DBName(), seedsTable)
	if err = cockroach.CloseDBOnError(db, err); err != nil {
		return nil, fmt.Errorf("instantiating DB definition: %s", err)
	}
	return &Model{db: db}, nil
}

