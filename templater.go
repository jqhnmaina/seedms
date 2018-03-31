package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/fileutils"
)

const (
	flagHelp = "help"
	flagDest = "dest"
	flagName = "name"
	flagDesc = "desc"

	envGOPATH = "GOPATH"

	seedms            = "seedms"
	seedmsDescription = "seedmsDescription"
	seedmsPkg         = "github.com/tomogoma/seedms"
)

var (
	keyWords = []string{seedms, seedmsDescription, seedmsPkg}

	help = flag.Bool(flagHelp, false, "Print out this help message")

	dest = flag.String(
		flagDest,
		"",
		"The micro-service's package location relative to $GOPATH/src e.g. github.com/tomogoma/seedms",
	)

	nameRe = regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*$")
	name   = flag.String(
		flagName,
		"",
		fmt.Sprintf("The name of the new micro-service, should conform to \"%s\"", nameRe.String()),
	)

	desc = flag.String(
		flagDesc,
		"",
		"Brief description of the micro-service",
	)
)

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	err := validateFlags(*dest, *name, *desc)
	handleError(err)

	GOPATH := os.Getenv(envGOPATH)
	if GOPATH == "" {
		handleError(fmt.Errorf("please ensure %s environment variable is set", envGOPATH))
	}
	srcDir := path.Join(GOPATH, "src")

	destFolder := path.Join(srcDir, *dest)
	destIsEmpty, err := fileutils.IsEmpty(destFolder)
	if !os.IsNotExist(err) {
		handleError(err)
	}
	if err == nil && !destIsEmpty {
		handleError(fmt.Errorf("%s (%s) exists and is not empty", flagDest, destFolder))
	}

	sourceFolder := path.Join(srcDir, seedmsPkg)

	if err := fileutils.CopyDir(sourceFolder, destFolder); err != nil {
		handleError(fmt.Errorf("unable to copy template files: %v", err))
	}
	destTemplater := path.Join(destFolder, "templater.go")
	if err := os.Remove(destTemplater); err != nil {
		warnOnError(fmt.Errorf("unable to remove %s: %v", destTemplater, err))
	}

	msReadMeFile := path.Join(destFolder, "README_MS.MD")
	resReadMeFile := path.Join(destFolder, "README.MD")
	if err := fileutils.CopyFile(msReadMeFile, resReadMeFile); err != nil {
		warnOnError(fmt.Errorf("unabl to set readme file: manually move"+
			" %s to %s: %v", msReadMeFile, resReadMeFile, err))
	}
	if err := os.Remove(msReadMeFile); err != nil {
		warnOnError(fmt.Errorf("unable to remove %s: %v", msReadMeFile, err))
	}

	if err := refactorNames(destFolder, *dest, *name, *desc); err != nil {
		warnOnError(fmt.Errorf("unable to refactor project values to match passed flags: %v", err))
	}
}

func validateFlags(dest, name, desc string) error {

	if dest == "" {
		return fmt.Errorf("%s flag is required", flagDest)
	}
	if in(dest, keyWords...) {
		return fmt.Errorf("%s flag cannot contain any of %v", flagDest, keyWords)
	}

	if name == "" {
		return fmt.Errorf("%s flag is required", flagName)
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf("%s flag value (%s) does not conform to \"%s\"",
			flagName, name, nameRe)
	}
	if in(name, keyWords...) {
		return fmt.Errorf("%s flag cannot contain any of %v", flagName, keyWords)
	}

	if in(desc, keyWords...) {
		return fmt.Errorf("%s flag cannot contain any of %v", flagDesc, keyWords)
	}

	return nil
}

func refactorNames(destFolder, destPkg, msName, msDesc string) error {

	escapedSeedmsPkg := escape(seedmsPkg, "/", "\\")
	escapedDestPkg := escape(destPkg, "/", "\\")

	return filepath.Walk(destFolder, func(fName string, info os.FileInfo, err error) error {

		if err != nil {
			return errors.Newf("error walking through %s: %v", fName, err)
		}
		if info.IsDir() {
			return nil
		}

		rplcPKGCmd := fmt.Sprintf("s/%s/%s/g", escapedSeedmsPkg, escapedDestPkg)
		cmd := exec.Command("sed", "-i", rplcPKGCmd, fName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("replace package name in %s: %s: %v",
				fName, out, err)
		}

		rplcDescCmd := fmt.Sprintf("s/%s/%s/g", seedmsDescription, msDesc)
		cmd = exec.Command("sed", "-i", rplcDescCmd, fName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("replace micro-service description in %s: %s: %v",
				fName, out, err)
		}

		rplcNameCmd := fmt.Sprintf("s/%s/%s/g", seedms, msName)
		cmd = exec.Command("sed", "-i", rplcNameCmd, fName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("replace micro-service name in %s: %s: %v",
				fName, out, err)
		}

		return nil
	})
}

func in(s string, checks ...string) bool {
	for _, check := range checks {
		if strings.Contains(s, check) {
			return true
		}
	}
	return false
}

func escape(s string, val string, escChar string) string {
	parts := strings.Split(s, val)
	var rslt string
	newVal := fmt.Sprintf("%s%s", escChar, val)
	for _, part := range parts {
		rslt = fmt.Sprintf("%s%s%s", rslt, part, newVal)
	}
	if !strings.HasSuffix(s, val) {
		rslt = strings.TrimSuffix(rslt, newVal)
	}
	return rslt
}

func handleError(err error) {
	if err == nil {
		return
	}
	flag.PrintDefaults()
	log.Fatal(err)
}

func warnOnError(err error) {
	if err == nil {
		return
	}
	log.Print(err)
}
