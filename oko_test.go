package oko_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

const TEST_DIR = "e2e"

func TestOko(t *testing.T) {
	type Test struct {
		Name string
		T    func(*testing.T)
	}
	for _, tests := range [][]Test{
		{
			{"Init", okoInit},
			{"Download", okoDownload},
			{"Install Local", okoInstallLocal},
			{"Remove Local", okoRemoveLocal},
		},
		{
			{"Migrate", okoMigrate},
		},
	} {
		for _, test := range tests {
			t.Run(test.Name, test.T)
		}
		cleanupFiles()
	}
	cleanup()
}

func cleanup() {
	if err := os.RemoveAll(TEST_DIR); err != nil {
		panic(err)
	}
}

func cleanupFiles() {
	cleanup()
	if err := os.Mkdir(TEST_DIR, os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func init() {
	cmd := exec.Command("make", "build")
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	cleanupFiles()
}

func okoDownload(t *testing.T) {
	if out, err := run(t, "download"); err != nil {
		t.Fatal(string(out), err)
	}
}

func okoInit(t *testing.T) {
	args := []string{"init"}
	if out, err := run(t, args...); err != nil {
		t.Fatal(string(out), err)
	}
	if out, err := run(t, args...); err == nil {
		t.Fatal(string(out))
	}
}

func okoRemoveLocal(t *testing.T) {
	args := []string{"r", "local", "src"}
	if out, err := run(t, args...); err != nil {
		t.Fatal(string(out), err)
	}
	if out, err := run(t, args...); err == nil {
		t.Fatal(string(out))
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
	if out, err := run(t, args...); err == nil {
		t.Fatal(string(out))
	}
}

func okoMigrate(t *testing.T) {
	path, err := exec.LookPath("vessel")
	if err != nil {
		t.Skip(err)
	}

	vessel := exec.Command(path, "init")
	vessel.Dir = TEST_DIR
	if out, err := vessel.CombinedOutput(); err != nil {
		t.Fatal(string(out), err)
	}

	args := []string{"migrate", "--keep"}
	if out, err := run(t, args...); err != nil {
		t.Fatal(string(out), err)
	}
	if out, err := run(t, args...); err == nil {
		t.Fatal(string(out))
	}
}

func run(t *testing.T, args ...string) ([]byte, error) {
	cmd := exec.Command("./../oko", args...)
	cmd.Dir = TEST_DIR
	return cmd.CombinedOutput()
}
