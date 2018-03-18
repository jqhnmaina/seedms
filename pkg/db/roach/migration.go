package roach

import (
	"errors"
	"fmt"
	"github.com/tomogoma/crdb"
)

func (r *Roach) migrate(fromVersion, toVersion int) error {

	var err error
	r.db, err = crdb.TryConnect(r.dsn, r.db)
	if err != nil {
		return fmt.Errorf("connect to db: %v", err)
	}

	// TODO supported migration logic here e.g.
	//		if fromVersion == 0 && toVersion == 1 {
	//			if err := r.migrate0To1(); err != nil { // implement r.migrate0To1()
	//				return err
	//			}
	//			return r.setRunningVersionCurrent()
	//		}

	return errors.New("not supported")
}
