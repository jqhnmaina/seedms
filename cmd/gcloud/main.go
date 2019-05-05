package main

import (
	"net/http"

	"github.com/tomogoma/seedms/internal/bootstrap"
	"github.com/tomogoma/seedms/pkg/config"
	"github.com/tomogoma/seedms/pkg/logging/logrus"
	"google.golang.org/appengine"
)

func main() {

	config.DefaultConfDir("conf")
	log := &logrus.Wrapper{}
	deps := bootstrap.Instantiate(config.DefaultConfPath(), log)

	httpHandler := bootstrap.NewHttpHandler(log, deps)

	http.Handle("/", httpHandler)
	appengine.Main()
}
