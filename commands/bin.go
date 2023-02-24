package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/github"
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
	Options: []cmd.Option{
		{
			Name:     "didc",
			HasValue: false,
		},
	},
	Method: func(_ []string, options map[string]string) error {
		pkg, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return NewBinError(err)
		}

		if pkg.CompilerVersion == nil {
			return NewBinError(NewCompilerVersionNotFoundError())
		}
		version := *pkg.CompilerVersion

		goos := runtime.GOOS // TODO: improve w/ GOARCH
		switch goos {
		case "darwin":
			goos = "macos"
		case "linux":
			goos = "linux64"
		default:
			return NewBinError(NewUnsupportedRuntimeErrors(goos))
		}

		if err := tar.DownloadGz(
			fmt.Sprintf(
				"https://github.com/dfinity/motoko/releases/download/%s/motoko-%s-%s.tar.gz",
				version, goos, version,
			),
			fmt.Sprintf(".oko/bin/%s", version),
		); err != nil {
			return NewBinError(err)
		}

		if _, ok := options["didc"]; ok {
			url := "dfinity/candid"
			resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases", url))
			if err != nil {
				return NewBinError(err)
			}
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return NewBinError(err)
			}
			var releases []github.Release
			if err := json.Unmarshal(data, &releases); err != nil {
				return NewBinError(err)
			}
			if len(releases) == 0 {
				return NewBinError(github.NewReleasesNotFoundErrors(url))
			}

			resp, err = http.Get(fmt.Sprintf(
				"https://github.com/%s/releases/download/%s/didc-%s",
				url, releases[0].TagName, goos,
			))
			if err != nil {
				return NewBinError(err)
			}
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				return NewBinError(err)
			}
			if err := os.WriteFile(fmt.Sprintf(".oko/bin/%s/didc", version), data, os.ModePerm); err != nil {
				return NewBinError(err)
			}
		}

		return nil
	},
}

var BinShowCommand = cmd.Command{
	Name:    "show",
	Aliases: []string{"s"},
	Summary: "prints out the path to the bin dir",
	Method: func(args []string, options map[string]string) error {
		pkg, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return NewBinError(err)
		}

		if pkg.CompilerVersion == nil {
			return NewBinError(NewCompilerVersionNotFoundError())
		}
		version := *pkg.CompilerVersion
		fmt.Printf(".oko/bin/%s", version)
		return nil
	},
}

type BinError struct {
	Err error
}

func NewBinError(err error) *BinError {
	return &BinError{
		Err: err,
	}
}

func (e BinError) Error() string {
	return fmt.Sprintf("bin error: %s", e.Err)
}

type CompilerVersionNotFoundError struct{}

func NewCompilerVersionNotFoundError() *CompilerVersionNotFoundError {
	return &CompilerVersionNotFoundError{}
}

func (e CompilerVersionNotFoundError) Error() string {
	return "no compiler version specified"
}

type UnsupportedRuntimeErrors struct {
	GOOS string
}

func NewUnsupportedRuntimeErrors(goos string) *UnsupportedRuntimeErrors {
	return &UnsupportedRuntimeErrors{
		GOOS: goos,
	}
}

func (e UnsupportedRuntimeErrors) Error() string {
	return fmt.Sprintf("unsupported runtime: %s", e.GOOS)
}
