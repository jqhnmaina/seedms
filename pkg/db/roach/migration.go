package roach

import "fmt"

func migrate(fromVersion, toVersion int) error {
	return fmt.Errorf("migration from %d to %d not supported",
		fromVersion, toVersion)
}
