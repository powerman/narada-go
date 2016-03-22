package narada

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	tmpdir, err := ioutil.TempDir("", "test-narada.")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		log.Fatal(err)
	}
	setupTestDir()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = os.RemoveAll(tmpdir)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func setupTestDir() error {
	dirs := []struct {
		name string
		perm os.FileMode
	}{
		{"config", 0755},
		{"config/dir", 0755},
		{"config/log", 0755},
		{"var", 0755},
	}
	files := []struct {
		name    string
		content string
		perm    os.FileMode
	}{
		{"VERSION", "1.2.3+example-1234567890 \n", 0644},
		{"config/file", "REAL1", 0644},
		{"config/dir/file", "Real2\n", 0644},
		{"config/unreadable", "", 0},
		{"config/empty", "", 0644},
		{"config/int", " 42 \n\n\n", 0644},
		{"config/badint", "42a", 0644},
		{"config/twoint", "42 777", 0644},
		{"config/float", "42.777", 0644},
		{"config/multi_line", "line1\n\nline2\n\n\n", 0644},
		{"config/single_line", "line1\n\n\n", 0644},
		{"config/duration", "3s", 0644},
		{"config/log/level", "INFO\n", 0644},
		{"config/log/output", "var/log.sock", 0644},
		{"config/log/type", "syslog", 0644},
	}
	for _, dir := range dirs {
		err := os.Mkdir(dir.name, dir.perm)
		if err != nil {
			return err
		}
	}
	for _, file := range files {
		err := ioutil.WriteFile(file.name, []byte(file.content), file.perm)
		if err != nil {
			return err
		}
	}
	return nil
}
