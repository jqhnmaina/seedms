package main

import (
	"net/http"

	"github.com/tomogoma/seedms/pkg/bootstrap"
	"github.com/tomogoma/seedms/pkg/config"
	httpInternal "github.com/tomogoma/seedms/pkg/handler/http"
	"github.com/tomogoma/seedms/pkg/logging"
	"github.com/tomogoma/seedms/pkg/logging/logrus"
	"google.golang.org/appengine"
)

func main() {

	config.DefaultConfDir("conf")
	log := &logrus.Wrapper{}
	deps := bootstrap.Instantiate(config.DefaultConfPath(), log)

	httpHandler, err := httpInternal.NewHandler(deps.Guard, log, config.WebRootPath(),
		deps.Config.Service.AllowedOrigins)
	logging.LogFatalOnError(log, err, "Instantiate http Handler")

	http.Handle("/", httpHandler)
	appengine.Main()
}
