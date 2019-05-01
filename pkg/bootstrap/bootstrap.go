package bootstrap

import (
	"github.com/tomogoma/seedms/pkg/db/gorm"
	"io/ioutil"

	"github.com/tomogoma/crdb"
	"github.com/tomogoma/go-api-guard"
	"github.com/tomogoma/jwt"
	"github.com/tomogoma/seedms/pkg/config"
	"github.com/tomogoma/seedms/pkg/db/roach"
	"github.com/tomogoma/seedms/pkg/logging"
)

type Deps struct {
	Config config.General
	Guard  *api.Guard
	Gorm   *gorm.Gorm
	JWTEr  *jwt.Handler
}

func InstantiateRoach(lg logging.Logger, conf crdb.Config) *roach.Roach {
	var opts []roach.Option
	if dsn := conf.FormatDSN(); dsn != "" {
		opts = append(opts, roach.WithDSN(dsn))
	}
	if dbn := conf.DBName; dbn != "" {
		opts = append(opts, roach.WithDBName(dbn))
	}
	rdb := roach.NewRoach(opts...)
	err := rdb.InitDBIfNot()
	logging.LogWarnOnError(lg, err, "Initiate Cockroach DB connection")
	return rdb
}

func InstantiateGorm(lg logging.Logger, conf crdb.Config) *gorm.Gorm {
	var opts []gorm.Option
	if dsn := conf.FormatDSN(); dsn != "" {
		opts = append(opts, gorm.WithDSN(dsn))
	}
	if dbn := conf.DBName; dbn != "" {
		opts = append(opts, gorm.WithDBName(dbn))
	}
	gormDB := gorm.NewGorm(opts...)
	err := gormDB.InitDBIfNot()
	logging.LogWarnOnError(lg, err, "Initiate Cockroach DB connection")
	return gormDB
}

func InstantiateJWTHandler(lg logging.Logger, tknKyF string) *jwt.Handler {
	JWTKey, err := ioutil.ReadFile(tknKyF)
	logging.LogFatalOnError(lg, err, "Read JWT key file")
	jwter, err := jwt.NewHandler(JWTKey)
	logging.LogFatalOnError(lg, err, "Instantiate JWT handler")
	return jwter
}

func Instantiate(confFile string, lg logging.Logger) Deps {

	conf, err := config.ReadFile(confFile)
	logging.LogFatalOnError(lg, err, "Read config file")

	gormDB := InstantiateGorm(lg, conf.Database)
	tg := InstantiateJWTHandler(lg, conf.Service.AuthTokenKeyFile)

	g, err := api.NewGuard(gormDB, api.WithMasterKey(conf.Service.MasterAPIKey))
	logging.LogFatalOnError(lg, err, "Instantate API access guard")

	return Deps{Config: conf, Guard: g, Gorm: gormDB, JWTEr: tg}
}
