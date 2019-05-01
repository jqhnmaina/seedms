package gorm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Roach is a cockroach db store.
// Use NewRoach() to instantiate.
type Gorm struct {
	errors.NotFoundErrCheck
	dsn              string
	dbName           string
	db               *gorm.DB
	compatibilityErr error

	isDBInitMutex sync.Mutex
	isDBInit      bool
}

const (
	keyDBVersion = "db.version"
	driverName   = "postgres"
)

// NewRoach creates an instance of *Roach. A db connection is only established
// when InitDBIfNot() or one of the Execute/Query methods is called.
func NewGorm(opts ...Option) *Gorm {
	g := &Gorm{
		isDBInit:      false,
		isDBInitMutex: sync.Mutex{},
		dbName:        config.CanonicalName(),
	}
	for _, f := range opts {
		f(g)
	}
	return g
}

// InitDBIfNot connects to and sets up the DB; creating it and tables if necessary.
func (gorm *Gorm) InitDBIfNot() error {
	var err error
	gorm.db, err = gorm.TryConnect()
	if err != nil {
		return errors.Newf("connect to db: %v", err)
	}
	return gorm.instantiate()
}

func (gorm *Gorm) TryConnect() (*gorm.DB, error) {
	db, err := gorm.Open(driverName, gorm.dsn)
	return db, err
}

// ExecuteTx prepares a transaction (with retries) for execution in fn.
// It commits the changes if fn returns nil, otherwise changes are rolled back.
//func (r *Roach) ExecuteTx(fn func(*sql.Tx) error) error {
//	if err := r.InitDBIfNot(); err != nil {
//		return err
//	}
//	return crdb.ExecuteTx(context.Background(), r.db, nil, fn)
//}

// ColDesc returns a string containing cols in the given order separated by ",".
func ColDesc(cols ...string) string {
	desc := ""
	for _, col := range cols {
		if col == "" {
			continue
		}
		desc = desc + col + ", "
	}
	return strings.TrimSuffix(desc, ", ")
}

func (gorm *Gorm) InstantiateDB(db *gorm.DB, dnName string) error {
	db.AutoMigrate(&ApiKey{}, &Configuration{})
	return nil
}

func (gorm *Gorm) instantiate() error {
	gorm.isDBInitMutex.Lock()
	defer gorm.isDBInitMutex.Unlock()
	if gorm.compatibilityErr != nil {
		return gorm.compatibilityErr
	}
	if gorm.isDBInit {
		return nil
	}
	if err := gorm.InstantiateDB(gorm.db, gorm.dbName); err != nil {
		return errors.Newf("instantiating db: %v", err)
	}
	if runningVersion, err := gorm.validateRunningVersion(); err != nil {
		if !gorm.IsNotFoundError(err) {
			if err != gorm.compatibilityErr {
				return fmt.Errorf("check db version: %v", err)
			}
			if err := gorm.migrate(runningVersion, Version); err != nil {
				return fmt.Errorf("migrate from version %d to %d: %v",
					runningVersion, Version, err)
			}
		}
		if err := gorm.setRunningVersionCurrent(); err != nil {
			return errors.Newf("set db version: %v", err)
		}
	}
	gorm.isDBInit = true
	return nil
}

func (gorm *Gorm) validateRunningVersion() (int, error) {
	var runningVersion int
	var confB Configuration
	if err := gorm.db.Where(ColKey+" = ?", keyDBVersion).First(&confB); err.Error != nil {
		if err.RecordNotFound() {
			return -1, errors.NewNotFoundf("config not found")
		}
		return -1, errors.Newf("get conf: %v", err.Error)
	}
	if err := json.Unmarshal([]byte(confB.Value), &runningVersion); err != nil {
		return -1, errors.Newf("Unmarshalling config: %v", err)
	}
	if runningVersion != Version {
		gorm.compatibilityErr = errors.Newf("db incompatible: need db"+
			" version '%d', found '%d'", Version, runningVersion)
		return runningVersion, gorm.compatibilityErr
	}
	return runningVersion, nil
}

func (gorm *Gorm) setRunningVersionCurrent() error {
	var dbVerConf Configuration
	gorm.db.Where(Configuration{Key: keyDBVersion}).Assign(Configuration{Value: strconv.Itoa(Version), UpdatedAt: time.Now()}).FirstOrCreate(&dbVerConf)
	if dbVerConf.Value == "" {
		return errors.Newf("unable to update db version")
	}
	gorm.compatibilityErr = nil
	return nil
}

func checkRowsAffected(r sql.Result, err error, expAffected int64) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewNotFound("none found")
		}
		return err
	}
	c, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.NewNotFound("none found for update")
	}
	if c != expAffected {
		return errors.Newf("expected %d affected rows but got %d",
			expAffected, c)
	}
	return nil
}
