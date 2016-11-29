package config

import (
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/auth/token"
	"time"
)

type Service struct {
	RegisterInterval time.Duration `yaml:"registerInterval,omitempty"`
}

type Config struct {
	Database cockroach.DSN `yaml:"database,omitempty"`
	Token    token.DefaultConfig `yaml:"token,omitempty"`
	Service  Service `yaml:"service,omitempty"`
}
