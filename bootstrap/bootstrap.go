package bootstrap

import (
	"io/ioutil"

	"github.com/tomogoma/go-api-guard"
	"github.com/tomogoma/jwt"
	"github.com/tomogoma/seedms/config"
	"github.com/tomogoma/seedms/db/roach"
	"github.com/tomogoma/seedms/logging"
	"github.com/tomogoma/crdb"
)

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

func InstantiateJWTHandler(lg logging.Logger, tknKyF string) *jwt.Handler {
	JWTKey, err := ioutil.ReadFile(tknKyF)
	logging.LogFatalOnError(lg, err, "Read JWT key file")
	jwter, err := jwt.NewHandler(JWTKey)
	logging.LogFatalOnError(lg, err, "Instantiate JWT handler")
	return jwter
}

func Instantiate(confFile string, lg logging.Logger) (config.General, *api.Guard, *roach.Roach, *jwt.Handler) {

	conf, err := config.ReadFile(confFile)
	logging.LogFatalOnError(lg, err, "Read config file")

	rdb := InstantiateRoach(lg, conf.Database)
	tg := InstantiateJWTHandler(lg, conf.Token.TokenKeyFile)

	g, err := api.NewGuard(rdb, api.WithMasterKey(conf.Service.MasterAPIKey))
	logging.LogFatalOnError(lg, err, "Instantate API access guard")

	return conf, g, rdb, tg
}
