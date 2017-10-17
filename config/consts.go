package config

import (
	"path"
	"time"
)

// Compile time constants that should not be configurable
// during runtime.
const (
	Name                = "seedms"
	Version             = "v0"
	Description         = "Seed Micro-Service"
	CanonicalName       = Name + Version
	RPCNamePrefix       = ""
	CanonicalRPCName    = RPCNamePrefix + CanonicalName
	WebNamePrefix       = "go.micro.api." + Version + "."
	WebRootURL          = "/" + Version + "/" + Name
	CanonicalWebName    = WebNamePrefix + Name
	DefaultSysDUnitName = CanonicalName + ".service"

	TimeFormat = time.RFC3339
)

var (
	// FIXME Probably won't work for none-unix systems!
	defaultInstallDir       = path.Join("/usr", "local", "bin")
	defaultSysDUnitFilePath = path.Join("/etc", "systemd", "system", DefaultSysDUnitName)
	sysDConfDir             = path.Join("/etc", Name)
	defaultConfDir          = sysDConfDir
)

func DefaultInstallDir() string {
	return defaultInstallDir
}

func DefaultInstallPath() string {
	return path.Join(defaultInstallDir, CanonicalName)
}

func DefaultSysDUnitFilePath() string {
	return defaultSysDUnitFilePath
}

func SysDConfDir() string {
	return sysDConfDir
}

// DefaultConfDir sets the value of the conf dir to use and returns it.
// It falls back to default - sysDConfDir - if newPSegments has zero len.
func DefaultConfDir(newPSegments ...string) string {
	if len(newPSegments) == 0 {
		defaultConfDir = sysDConfDir
	} else {
		defaultConfDir = path.Join(newPSegments...)
	}
	return defaultConfDir
}

func DefaultConfPath() string {
	return path.Join(defaultConfDir, CanonicalName+".conf.yml")
}
