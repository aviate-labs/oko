package commands

import (
	"fmt"
	"runtime"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
	"github.com/internet-computer/oko/internal/tar"
)

var BinCommand = cmd.Command{
	Name:    "bin",
	Aliases: []string{"b"},
	Summary: "Motoko compiler stuff",
	Commands: []cmd.Command{
		BinDownloadCommand,
		BinShowCommand,
	},
}

var BinDownloadCommand = cmd.Command{
	Name:    "download",
	Aliases: []string{"d"},
	Summary: "downloads the Motoko compiler",
	Method: func(_ []string, _ map[string]string) error {
		pkg, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return fmt.Errorf("`oko.json` already exists: %s", err)
		}

		if pkg.CompilerVersion == nil {
			return fmt.Errorf("no compiler version specified in `oko.json`")
		}
		version := *pkg.CompilerVersion

		goos := runtime.GOOS // TODO: improve w/ GOARCH
		switch goos {
		case "darwin":
			goos = "macos"
		case "linux":
			goos = "linux64"
		default:
			return fmt.Errorf("unsupported runtime: %s", goos)
		}

		return tar.DownloadGz(
			fmt.Sprintf(
				"https://github.com/dfinity/motoko/releases/download/%s/motoko-%s-%s.tar.gz",
				version, goos, version,
			),
			fmt.Sprintf(".oko/bin/%s", version),
		)
	},
}

var BinShowCommand = cmd.Command{
	Name:    "show",
	Aliases: []string{"s"},
	Summary: "prints out the path to the bin dir",
	Method: func(args []string, options map[string]string) error {
		pkg, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return fmt.Errorf("`oko.json` does not exist: %s", err)
		}

		if pkg.CompilerVersion == nil {
			return fmt.Errorf("no compiler version specified in `oko.json`")
		}
		version := *pkg.CompilerVersion
		fmt.Printf(".oko/bin/%s", version)
		return nil
	},
}
