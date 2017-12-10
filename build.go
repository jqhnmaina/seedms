//// +build dev

// build.go automates proper versioning of authms binaries
// and installer scripts.
// Use it like:   go run build.go
// The result binary will be located in bin/app
// You can customize the build with the -goos, -goarch, and
// -goarm CLI options:   go run build.go -goos=windows
//
// This program is NOT required to build authms from source
// since it is go-gettable. (You can run plain `go build`
// in this directory to get a binary).
package main

import (
	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/config"
	"flag"
	"os"
	"io/ioutil"
	"path"
	"os/exec"
	"bytes"
	"io"
	"path/filepath"
	"log"
	"encoding/json"
	"fmt"
)

func main() {
	var goos, goarch, goarm string
	var help bool
	flag.StringVar(&goos, "goos", "",
		"GOOS\tThe operating system for which to compile\n"+
			"\t\tExamples are linux, darwin, windows, netbsd.")
	flag.StringVar(&goarch, "goarch", "",
		"GOARCH\tThe architecture, or processor, for which to compile code.\n"+
			"\t\tExamples are amd64, 386, arm, ppc64.")
	flag.StringVar(&goarm, "goarm", "",
		"GOARM\tFor GOARCH=arm, the ARM architecture for which to compile.\n"+
			"\t\tValid values are 5, 6, 7.")
	flag.BoolVar(&help, "help", false, "Show this help message")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	if err := buildMicroservice(goos, goarch, goarm); err != nil {
		log.Fatalf("buildMicroservice error: %v", err)
	}
	if err := installVars(); err != nil {
		log.Fatalf("write installer script error: %v", err)
	}
	if err := buildGcloud(); err != nil {
		log.Fatalf("build GCloud error: %v", err)
	}
}

func installVars() error {
	content := `#!/usr/bin/env bash
NAME="` + config.Name + `"
VERSION="` + config.VersionFull + `"
DESCRIPTION="` + config.Description + `"
CANONICAL_NAME="` + config.CanonicalName() + `"
CONF_DIR="` + config.DefaultConfDir() + `"
CONF_FILE="` + config.DefaultConfPath() + `"
INSTALL_DIR="` + config.DefaultInstallDir() + `"
INSTALL_FILE="` + config.DefaultInstallPath() + `"
UNIT_NAME="` + config.DefaultSysDUnitName() + `"
UNIT_FILE="` + config.DefaultSysDUnitFilePath() + `"
DOCS_DIR="` + config.DefaultDocsDir() + `"
`
	return ioutil.WriteFile("install/vars.sh", []byte(content), 0755)
}

func buildMicroservice(goos, goarch, goarm string) error {
	docsDir := path.Join("install", "docs", config.VersionMajorPrefixed(), config.Name, "docs")
	if err := compileDocs(docsDir); err != nil {
		return err
	}
	args := []string{"build", "-o", "bin/app", "./cmd/micro"}
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()
	for _, env := range []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		"GOARM=" + goarm,
	} {
		cmd.Env = append(cmd.Env, env)
	}
	// TODO USE cmd.CombinedOutput()
	return cmd.Run()
}

func buildGcloud() error {
	confDir := config.DefaultConfDir("cmd", "gcloud", "conf")

	if err := os.MkdirAll(confDir, 0755); err != nil {
		return errors.Newf("create conf dir: %v", err)
	}

	docsDir := path.Join(config.DefaultDocsDir(), config.VersionMajorPrefixed(), config.Name, "docs")
	if err := compileDocs(docsDir); err != nil {
		return err
	}

	err := copyIfDestNotExists(path.Join("install", "conf.yml"), config.DefaultConfPath())
	if err != nil {
		return err
	}
	if err := cleanGCloudConfFile(); err != nil {
		return errors.Newf("clean gcloud config file: %v", err)
	}

	return nil
}

func compileDocs(docsDir string) error {

	subjDir := path.Join("handler", "http")
	headerFile := path.Join(subjDir, "apidoc_header.md")
	APIDocConfFile := path.Join(subjDir, "apidoc.json")

	apiDoc := struct {
		Name        string      `json:"name"`
		Version     string      `json:"version"`
		Description string      `json:"description"`
		Title       string      `json:"title"`
		Header      interface{} `json:"header"`
	}{
		Name:        config.Name,
		Version:     config.VersionFull,
		Description: config.Description,
		Title:       config.CanonicalName(),
		Header: struct {
			Title    string `json:"title"`
			FileName string `json:"filename"`
		}{
			Title:    "Introduction",
			FileName: headerFile,
		},
	}

	apiDocB, err := json.Marshal(apiDoc)
	if err != nil {
		return errors.Newf("Marshal API doc config: %v", err)
	}

	err = ioutil.WriteFile(APIDocConfFile, apiDocB, 0655)
	if err != nil {
		return errors.Newf("Write API doc file: %v", err)
	}

	args := []string{"-i", subjDir, "-c", subjDir, "-o", docsDir}
	cmd := exec.Command("apidoc", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Newf("generate http docs: %v: %s", err, out)
	}
	return nil
}

func cleanGCloudConfFile() error {
	newPath := config.DefaultConfPath()
	confContent, err := ioutil.ReadFile(config.DefaultConfPath())
	if err != nil {
		return errors.Newf("read file for transform: %v", err)
	}
	confContentClean := bytes.Replace(confContent, []byte(config.SysDConfDir()+"/"), []byte("conf/"), -1)
	err = ioutil.WriteFile(newPath, confContentClean, 0644)
	if err != nil {
		return errors.Newf("write transformed file: %v", err)
	}
	return nil
}

// copyIfDestNotExists copyis the from file into dest file if dest does not exists.
// see copyFile for Notes.
func copyIfDestNotExists(from, dest string) error {
	_, err := os.Stat(dest)
	if err == nil {
		fmt.Printf("'%s' ignored, already exists\n", dest)
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.Newf("stat: %v", err)
	}
	return copyFile(from, dest)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return errors.Newf("open src: %v", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return errors.Newf("create dst file: %v", err)
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = errors.Newf("%v ...close dst file: %v", err, e)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Newf("copy contents: %v", err)
	}

	err = out.Sync()
	if err != nil {
		return errors.Newf("flush buffer after copy: %v", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return errors.Newf("stat src file: %v", err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return errors.Newf("chmod dest to equal src perms: %v", err)
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist.
// Any source content that already exists in destination will be ignored and skipped.
// Symlinks are ignored and skipped.
func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return errors.Newf("stat source: %v", err)
	}
	if !si.IsDir() {
		return errors.Newf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return errors.Newf("stat destination: %v", err)
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return errors.Newf("mkdirall destination: %v", err)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.Newf("read source: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return errors.Newf("copy child dir: %v", err)
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyIfDestNotExists(srcPath, dstPath)
			if err != nil {
				return errors.Newf("copy child file: %v", err)
			}
		}
	}

	return
}
