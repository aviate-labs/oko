package commands_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/internet-computer/oko/commands"
)

const TEST_DIR = "e2e"

func TestCommands(t *testing.T) {
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
}

func cleanupFiles() {
	_ = os.Remove("./oko.json")
	_ = os.RemoveAll("./src")
	_ = os.Remove("./vessel.dhall")
	_ = os.Remove("./package-set.dhall")
}

func okoDownload(t *testing.T) {
	if err := commands.DownloadCommand.Call(); err != nil {
		t.Fatal(err)
	}
}

func okoInit(t *testing.T) {
	if err := commands.InitCommand.Call(); err != nil {
		t.Fatal(err)
	}
	if err := commands.InitCommand.Call(); err == nil {
		t.Fatal()
	}
}

func okoInstallLocal(t *testing.T) {
	args := []string{"local", "src", "--name=src"}
	if err := commands.InstallCommand.Call(args...); err == nil {
		t.Fatal()
	}
	if err := os.Mkdir("src", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := commands.InstallCommand.Call(args...); err != nil {
		t.Fatal(err)
	}
	if err := commands.InstallCommand.Call(args...); err == nil {
		t.Fatal()
	}
}

func okoMigrate(t *testing.T) {
	path, err := exec.LookPath("vessel")
	if err != nil {
		t.Skip(err)
	}

	vessel := exec.Command(path, "init")
	if out, err := vessel.CombinedOutput(); err != nil {
		t.Fatal(string(out), err)
	}

	args := []string{"--keep"}
	if err := commands.MigrateCommand.Call(args...); err != nil {
		t.Fatal(err)
	}
	if err := commands.MigrateCommand.Call(args...); err == nil {
		t.Fatal()
	}
}

func okoRemoveLocal(t *testing.T) {
	args := []string{"src"}
	if err := commands.RemoveCommand.Call(args...); err != nil {
		t.Fatal(err)
	}
	if err := commands.RemoveCommand.Call(args...); err == nil {
		t.Fatal()
	}
}

type Test struct {
	Name string
	T    func(*testing.T)
}
