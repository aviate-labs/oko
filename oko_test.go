package oko_test

import (
	"os"
	"os/exec"
	"testing"
)

const TEST_DIR = "e2e"

func TestOko(t *testing.T) {
	t.Run("Init", okoInit)
	t.Run("Download", okoDownload)

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
	if _, err := run(t, "init"); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "init"); err == nil {
		t.Error(string(out))
	}
}

func run(t *testing.T, args ...string) ([]byte, error) {
	cmd := exec.Command("./../oko", args...)
	cmd.Dir = TEST_DIR
	return cmd.CombinedOutput()
}
