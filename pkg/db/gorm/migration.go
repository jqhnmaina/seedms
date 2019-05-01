package gorm

import (
	"errors"
	"fmt"
)

func (r *Gorm) migrate(fromVersion, toVersion int) error {

	var err error
	r.db, err = r.TryConnect()
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
