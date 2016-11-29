package main

import (
	"flag"
	"time"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	// TODO SEEDMS fix imports after renaming project root folder
	// (replace all "github.com/tomogoma/seedms" refs with new path)
	"github.com/tomogoma/seedms/server"
	"github.com/tomogoma/seedms/model"
	"github.com/tomogoma/seedms/config"
	"github.com/tomogoma/seedms/server/proto"
	"github.com/tomogoma/go-commons/auth/token"
)

const (
	// TODO SEEDMS change the name of the micro-service to a desired value
	// (preferably the same as the NAME value in install/systemd-install.sh)
	name = "seedms"
	version = "0.1.0"
	confCommand = "conf"
	defaultConfFile = "/etc/" + name + "/" + name + ".conf.yml"
)

var confFilePath = flag.String(confCommand, defaultConfFile, "path to config file")

func main() {
	flag.Parse();
	defer func() {
		time.Sleep(600 * time.Millisecond)
	}()
	log := log4go.NewDefaultLogger(log4go.FINEST)
	conf, err := config.ReadFile(*confFilePath)
	if err != nil {
		log.Critical("Error reading config file: %s", err)
		return
	}
	m, err := model.New(conf.Database)
	if err != nil {
		log.Critical("Error instantiating the model: %s", err)
		return
	}
	tv, err := token.NewGenerator(conf.Token)
	if err != nil {
		log.Critical("Error instantiating token validator: %s", err)
		return
	}
	srv, err := server.New(name, m, tv, log);
	if err != nil {
		log.Critical("Error instantiating server: %s", err)
		return
	}
	service := micro.NewService(
		micro.Name(name),
		micro.Version(version),
		micro.RegisterInterval(conf.Service.RegisterInterval),
	)
	seed.RegisterSeedHandler(service.Server(), srv)
	if err := service.Run(); err != nil {
		log.Critical("Error serving: %s", err)
	}
}
