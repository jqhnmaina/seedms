package testing

import (
	"flag"
	"strconv"
	"sync"
	"testing"

	"github.com/tomogoma/seedms/config"
)

var confPath = flag.String("testConf", config.DefaultConfPath(), "/path/to/config")

var currIDMutex = sync.Mutex{}
var currID = 0

func init() {
	flag.Parse()
}

func currentID() string {
	currIDMutex.Lock()
	defer currIDMutex.Unlock()
	currID++
	return strconv.Itoa(currID)
}

func ReadConfig(t *testing.T) config.General {
	conf, err := config.ReadFile(*confPath)
	if err != nil {
		t.Fatalf("Error setting up: read config: %v", err)
	}
	return conf
}
