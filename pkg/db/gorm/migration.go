package gorm

import (
	"errors"
	"fmt"
)

func (gorm *Gorm) migrate(fromVersion, toVersion int) error {

	var err error
	gorm.db, err = gorm.TryConnect()
	if err != nil {
		return fmt.Errorf("connect to db: %v", err)
	}

	// TODO supported migration logic here e.gorm.
	//		if fromVersion == 0 && toVersion == 1 {
	//			if err := gorm.migrate0To1(); err != nil { // implement gorm.migrate0To1()
	//				return err
	//			}
	//			return gorm.setRunningVersionCurrent()
	//		}

	return errors.New("not supported")
}
