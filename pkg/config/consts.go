package config

import (
	"path"
	"strings"
)

// Compile time constants that should not be configurable
// during runtime.
const (
	Name        = "seedms"
	VersionFull = "0.1.0" // Use http://semver.org standards
	Description = "seedmsDescription"

	// be sure to run micro with the new namespace after changing this e.g.
	//     $micro api --namespace=new.namespace.value ...
	// or set the environment value.
	// Docs here: https://micro.mu/docs/api.html#set-namespace
	Namespace   = "go.micro.api"

	RPCNamePrefix = ""

	DocsPath = "docs"
)

var (
	// FIXME Probably won't work for none-unix systems!
	defaultInstallDir       = path.Join("/usr", "local", "bin")
	defaultSysDUnitFilePath = path.Join("/etc", "systemd", "system", DefaultSysDUnitName())
	sysDConfDir             = path.Join("/etc", Name)
	defaultConfDir          = sysDConfDir
)

func CanonicalName() string {
	return Name + VersionMajorPrefixed(VersionFull, "")
}

func CanonicalRPCName() string {
	return RPCNamePrefix + CanonicalName()
}

// VersionMajorPrefixed returns the semver major version in VersionFull if greater than zero otherwise returns
// the first of minor/patch with a non-zero value separated by sep e.g.
//    versionFull = "2.1.3", sep = "_" -> "v2"
//    versionFull = "0.4.2", sep = "_" -> "v0_4"
//    versionFull = "0.0.2", sep = "_" -> "v0_0_2"
// This is useful when defining URLs for services where the server treats dots (.) as special characters.
// Behaviour is undefined when versionFull does not follow semver 2.0.0 rules, but will probably
// default to returning "v0".
func VersionMajorPrefixed(versionFull, sep string) string {

	// remove any pre-release info e.g. "-alpha1" in "1.0.0-alpha1"
	preReleaseIdx := strings.Index(versionFull, "-")
	if preReleaseIdx != -1 {
		versionFull = versionFull[0:preReleaseIdx]
	}

	versionSplit := strings.Split(versionFull, ".")
	val := "v" + versionSplit[0] + sep

	if len(versionSplit) > 3 {
		versionSplit = versionSplit[0:3]
	}

	for i := 1; versionSplit[i-1] == "0" && i < len(versionSplit); i = i + 1 {
		val = val + versionSplit[i] + sep
	}

	val = strings.TrimSuffix(val, sep)
	if val == "v" || strings.HasSuffix(val, "0") {
		val = "v0"
	}
	return val
}

func WebNamePrefix() string {
	return Namespace + "." + VersionMajorPrefixed(VersionFull, "") + "."
}

func WebRootPath() string {
	return "/" + VersionMajorPrefixed(VersionFull, "") + "/" + Name
}

func CanonicalWebName() string {
	return WebNamePrefix() + Name
}

func DefaultSysDUnitName() string {
	return CanonicalName() + ".service"
}

func DefaultInstallDir() string {
	return defaultInstallDir
}

func DefaultInstallPath() string {
	return path.Join(defaultInstallDir, CanonicalName())
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

func DefaultDocsDir() string {
	return path.Join(defaultConfDir, DocsPath)
}

func DefaultConfPath() string {
	return path.Join(defaultConfDir, CanonicalName()+".conf.yml")
}
