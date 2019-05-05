package bootstrap

import (
	"github.com/tomogoma/crdb"
	"github.com/tomogoma/go-api-guard"
	"github.com/tomogoma/jwt"
	"github.com/tomogoma/seedms/pkg/config"
	"github.com/tomogoma/seedms/pkg/db/gorm"
	httpApi "github.com/tomogoma/seedms/pkg/handler/http"
	"github.com/tomogoma/seedms/pkg/handler/http/status"
	jwtH "github.com/tomogoma/seedms/pkg/jwt"
	"github.com/tomogoma/seedms/pkg/logging"
	"io/ioutil"
	"net/http"
)

type Deps struct {
	Config config.General
	Guard  *api.Guard
	Gorm   *gorm.Gorm
	JWTEr  *jwt.Handler
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

func NewStatusSubRoute() httpApi.SubRoute {
	sh := status.NewHandler()
	return httpApi.SubRoute{Path: "/status", Handler: sh}
}

func NewHttpHandler(lg logging.Logger, deps Deps) http.Handler {
	JWTKey, err := ioutil.ReadFile(deps.Config.Service.AuthTokenKeyFile)
	logging.LogFatalOnError(lg, err, "Read JWT key file")

	jwtHandler, err := jwt.NewHandler(JWTKey)
	logging.LogFatalOnError(lg, err, "Instantiate JWT handler")

	jwtHelper, err := jwtH.NewHelper(jwtHandler)
	logging.LogFatalOnError(lg, err, "Instantiate JWT helper")

	h, err := httpApi.NewHandler(deps.Guard, config.WebRootPath(), lg, jwtHelper, deps.Config.Service.DocsDir, deps.Config.Service.AllowedOrigins,
		NewStatusSubRoute(),
	)
	logging.LogFatalOnError(lg, err, "Instantiate http API handler")

	return h
}
