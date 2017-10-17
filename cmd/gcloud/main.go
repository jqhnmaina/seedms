package main

import (
	"net/http"

	"github.com/tomogoma/seedms/bootstrap"
	"github.com/tomogoma/seedms/config"
	httpInternal "github.com/tomogoma/seedms/handler/http"
	"github.com/tomogoma/seedms/logging"
	"github.com/tomogoma/seedms/logging/logrus"
	"google.golang.org/appengine"
)

func main() {

	config.DefaultConfDir("conf")
	log := &logrus.Wrapper{}
	_, APIGuard, _, _ := bootstrap.Instantiate(config.DefaultConfPath(), log)

	httpHandler, err := httpInternal.NewHandler(APIGuard, log)
	logging.LogFatalOnError(log, err, "Instantiate http Handler")

	http.Handle("/", httpHandler)
	appengine.Main()
}
