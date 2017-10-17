package main

import (
	"flag"
	http2 "net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/seedms/api"
	"github.com/tomogoma/seedms/bootstrap"
	"github.com/tomogoma/seedms/config"
	"github.com/tomogoma/seedms/handler/http"
	"github.com/tomogoma/seedms/handler/rpc"
	"github.com/tomogoma/seedms/logging"
	"github.com/tomogoma/seedms/logging/logrus"
	_ "github.com/tomogoma/seedms/logging/standard"
)

func main() {

	confFile := flag.String("conf", config.DefaultConfPath(), "location of config file")
	flag.Parse()
	log := &logrus.Wrapper{}
	conf, APIGuard, _, _ := bootstrap.Instantiate(*confFile, log)

	serverRPCQuitCh := make(chan error)
	rpcSrv, err := rpc.NewStatusHandler(APIGuard, log)
	logging.LogFatalOnError(log, err, "Instantate RPC handler")
	go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)

	serverHttpQuitCh := make(chan error)
	httpHandler, err := http.NewHandler(APIGuard, log)
	logging.LogFatalOnError(log, err, "Instantiate HTTP handler")
	go serveHttp(conf.Service, httpHandler, serverHttpQuitCh)

	select {
	case err = <-serverHttpQuitCh:
		logging.LogFatalOnError(log, err, "Serve HTTP")
	case err = <-serverRPCQuitCh:
		logging.LogFatalOnError(log, err, "Serve RPC")
	}
}

func serveRPC(conf config.Service, rpcSrv *rpc.StatusHandler, quitCh chan error) {
	service := micro.NewService(
		micro.Name(config.CanonicalRPCName),
		micro.Version(conf.LoadBalanceVersion),
		micro.RegisterInterval(conf.RegisterInterval),
	)
	api.RegisterStatusHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
}

func serveHttp(conf config.Service, h http2.Handler, quitCh chan error) {
	srvc := web.NewService(
		web.Handler(h),
		web.Name(config.CanonicalWebName),
		web.Version(conf.LoadBalanceVersion),
		web.RegisterInterval(conf.RegisterInterval),
	)
	quitCh <- srvc.Run()
}
