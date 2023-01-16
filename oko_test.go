package oko_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

const TEST_DIR = "e2e"

func TestOko(t *testing.T) {
	for _, v := range []struct {
		Name string
		T    func(*testing.T)
	}{
		{"Init", okoInit},
		{"Download", okoDownload},
		{"Install Local", okoInstallLocal},
	} {
		t.Run(v.Name, v.T)
	}

	cleanup()
}

func cleanup() {
	if err := os.RemoveAll(TEST_DIR); err != nil {
		panic(err)
	}
}

func init() {
	cmd := exec.Command("make")
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	cleanup()
	if err := os.Mkdir(TEST_DIR, os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func okoDownload(t *testing.T) {
	if _, err := run(t, "download"); err != nil {
		t.Fatal(err)
	}
}

func okoInit(t *testing.T) {
	args := []string{"init"}
	if _, err := run(t, args...); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, args...); err == nil {
		t.Error(string(out))
	}
}

func okoInstallLocal(t *testing.T) {
	args := []string{"i", "local", "src", "--name=src"}
	if out, err := run(t, args...); err == nil {
		t.Fatal(string(out))
	}
	if err := os.Mkdir(fmt.Sprintf("%s/src", TEST_DIR), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, args...); err != nil {
		t.Fatal(string(out), err)
	}
}

func run(t *testing.T, args ...string) ([]byte, error) {
	cmd := exec.Command("./../oko", args...)
	cmd.Dir = TEST_DIR
	return cmd.CombinedOutput()
}
