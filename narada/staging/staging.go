// Package staging provides temporary Narada project directory for tests.
// It must be imported in first import statement by all files of your package
// (including non-test files) which imports any narada-aware package,
// even before narada/bootstrap import in main package.
//
// Importing this package will have effect only under `go test`:
// current directory will be changed to temporary Narada project directory.
// To cleanup that directory after tests call TearDown like this:
//
//   func TestMain(m *testing.M) { os.Exit(staging.TearDown(m.Run())) }
package staging

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	pkgDir string
	tmpDir string
	dirs   = []string{
		".backup",
		"config",
		"config/backup",
		"config/log",
		"config/mysql",
		"config/mysql/dump",
		"config/qmail",
		"tmp",
		"var",
		"var/log",
		"var/use",
		"var/mysql",
		"var/qmail",
	}
	files = []struct{ name, data string }{
		{"config/backup/exclude", "./.backup/*\n./.lock*\n./tmp/*\n./.release/*\n"},
		{"config/log/level", "DEBUG"},
		{"config/log/type", "file"},
		{"config/log/file", "/dev/stdout"},
		{"config/mysql/host", ""},
		{"config/mysql/port", "3306"},
		{"config/mysql/db", ""},
		{"config/mysql/login", ""},
		{"config/mysql/pass", ""},
		{"config/mysql/dump/empty", ""},
		{"config/mysql/dump/ignore", ""},
		{"config/mysql/dump/incremental", ""},
	}
)

func init() {
	if flag.Lookup("test.v") != nil {
		err := setUp()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setUp() (err error) {
	pkgDir, err = os.Getwd()
	if err != nil {
		return err
	}
	tmpDir, err = ioutil.TempDir("", "narada-staging.")
	if err != nil {
		return err
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return err
		}
	}
	for _, file := range files {
		err = ioutil.WriteFile(file.name, []byte(file.data), 0666)
		if err != nil {
			return err
		}
	}

	custom := pkgDir + "/testdata/staging.setup"
	if _, err = os.Stat(custom); err == nil {
		cmd := exec.Command(custom, pkgDir)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func TearDown(exitCode int) int {
	var err error
	custom := pkgDir + "/testdata/staging.teardown"
	if _, err = os.Stat(custom); err == nil {
		cmd := exec.Command(custom, pkgDir)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Print(err)
			return exitCode
		}
	}

	err = os.Chdir(pkgDir)
	if err != nil {
		log.Print(err)
		return exitCode
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		log.Print(err)
		return exitCode
	}
	return exitCode
}
