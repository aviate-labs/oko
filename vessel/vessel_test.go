package vessel_test

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"testing"

	"github.com/internet-computer/oko/vessel"
)

const TEST_DIR = "e2e"

func cleanup() {
	if err := os.RemoveAll(TEST_DIR); err != nil {
		panic(err)
	}
}

func init() {
	cleanup()
	if err := os.Mkdir(TEST_DIR, os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func TestVessel(t *testing.T) {
	path, err := exec.LookPath("vessel")
	if err != nil {
		t.Skip()
	}
	cmd := exec.Command(path, "init")
	cmd.Dir = TEST_DIR
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Error(err, string(out))
	}

	if _, err := vessel.LoadManifest(fmt.Sprintf("%s/vessel.dhall", TEST_DIR)); err != nil {
		t.Error(err)
	}

	if _, err := vessel.LoadPackageSet(fmt.Sprintf("%s/package-set.dhall", TEST_DIR)); err != nil {
		switch err := err.(type) {
		case *url.Error:
			switch err := err.Err.(type) {
			case *net.OpError:
				t.Skip(err)
			default:
				t.Error(err)
			}
		default:
			t.Error(err)
		}
	}

	cleanup()
}
