package gorm

import (
	"errors"
	"fmt"
)

func (g *Gorm) migrate(fromVersion, toVersion int) error {

	var err error
	g.db, err = g.TryConnect()
	if err != nil {
		return fmt.Errorf("connect to db: %v", err)
	}

	// TODO supported migration logic here e.g.
	//		if fromVersion == 0 && toVersion == 1 {
	//			if err := g.migrate0To1(); err != nil { // implement g.migrate0To1()
	//				return err
	//			}
	//			return g.setRunningVersionCurrent()
	//		}

	return errors.New("not supported")
}
