package main

import (
	"flag"
	"time"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	// TODO SEEDMS fix imports after renaming project root folder
	// (replace all "github.com/tomogoma/seedms" refs with new path)
	"github.com/tomogoma/seedms/handler"
	"github.com/tomogoma/seedms/handler/proto"
	"github.com/tomogoma/go-commons/auth/token"
	confhelper "github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/seedms/config"
	"runtime"
	"fmt"
)

const (
	// TODO SEEDMS change the name of the micro-service to a desired value
	// (preferably the same as the NAME value in install/systemd-install.sh)
	name = "seedms"
	apiID = "go.micro.api." + name
	version = "0.1.0"
	confCommand = "conf"
	defaultConfFile = "/etc/" + name + "/" + name + ".conf.yaml"
)

type Logger interface {
	Fine(interface{}, ...interface{})
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

var confFilePath = flag.String(confCommand, defaultConfFile, "path to config file")

func main() {
	flag.Parse()
	defer func() {
		runtime.Gosched()
		time.Sleep(50 * time.Millisecond)
	}()
	conf := config.Config{}
	log := log4go.NewDefaultLogger(log4go.FINEST)
	err := confhelper.ReadYamlConfig(*confFilePath, &conf)
	if err != nil {
		log.Critical("Error reading config file: %s", err)
		return
	}
	err = bootstrap(log, conf)
	log.Critical(err)
}

// bootstrap collects all the dependencies necessary to start the server,
// injects said dependencies, and proceeds to register it as a micro grpc handler.
func bootstrap(log Logger, conf config.Config) error {
	tv, err := token.NewGenerator(conf.Auth)
	if err != nil {
		return fmt.Errorf("Error instantiating token validator: %s", err)
	}
	seedH, err := handler.NewSeed(apiID, tv, log);
	if err != nil {
		return fmt.Errorf("Error instantiating server: %s", err)
	}
	service := micro.NewService(
		micro.Name(apiID),
		micro.Version(version),
		micro.RegisterInterval(conf.Service.RegisterInterval),
	)
	// TODO SEEDMS modify this to match .proto file specification
	proto.RegisterSeedHandler(service.Server(), seedH)
	if err := service.Run(); err != nil {
		return fmt.Errorf("Error serving: %s", err)
	}
	return nil
}
