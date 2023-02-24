package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/github"
	"github.com/internet-computer/oko/internal/cmd"
	"github.com/internet-computer/oko/vessel"
)

var InstallCommand = cmd.Command{
	Name:        "install",
	Aliases:     []string{"i"},
	Summary:     "install packages",
	Description: `Allows you to install packages from GitHub or link local directories.`,
	Commands: []cmd.Command{
		InstallGitHubCommand,
		InstallLocalCommand,
	},
}

var InstallGitHubCommand = cmd.Command{
	Name:    "github",
	Aliases: []string{"gh"},
	Summary: "install GitHub hosted packages",
	Description: "Allows you to install packages from GitHub.\n\n" +
		"Expects `{org}/{repo}`, i.e. if you want to install the package at https://github.com/internet-computer/testing.mo you will have to pass `internet-computer/testing.mo` to the first argument.\n\n" +
		"Instead of specifying a specific version, `latest` can be used.",
	Args: []string{"url", "version"},
	Options: []cmd.Option{
		{
			Name:     "name",
			Summary:  "package name",
			HasValue: true,
		},
	},
	Method: func(args []string, options map[string]string) error {
		url := args[0]
		version := args[1]
		if version == "latest" {
			resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases", url))
			if err != nil {
				return NewInstallError(err)
			}
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return NewInstallError(err)
			}
			var releases []github.Release
			if err := json.Unmarshal(data, &releases); err != nil {
				return NewInstallError(err)
			}
			if len(releases) == 0 {
				return NewInstallError(github.NewReleasesNotFoundErrors(url))
			}
			version = releases[0].TagName
		}

		info := config.PackageInfoRemote{
			Repository: fmt.Sprintf("https://github.com/%s", url),
			Version:    version,
		}

		name, ok := options["name"]
		if !ok {
			// Ask for rename of package.
			name := url[strings.LastIndex(url, "/")+1:]
			if cmd.AskForConfirmation(fmt.Sprintf("Do you want to rename the package name %q?", name)) {
				info.Name = strings.TrimSpace(cmd.Ask("New name"))
			} else {
				info.Name = name
			}
		} else {
			info.Name = name
		}

		state, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return NewInstallError(err)
		}

		if err := info.Download(); err != nil {
			return NewInstallError(err)
		}

		// VESSEL
		if raw, err := os.ReadFile(fmt.Sprintf("%s/vessel.dhall", info.RelativePath())); err == nil {
			manifest, err := vessel.NewManifest(raw)
			if err != nil {
				return NewInstallError(err)
			}
			info.Dependencies = manifest.Dependencies
			if len(manifest.Dependencies) != 0 {
				packageSet, err := vessel.LoadPackageSet(fmt.Sprintf("%s/package-set.dhall", info.RelativePath()))
				if err != nil {
					return NewInstallError(err)
				}
				packages, err := packageSet.Filter(manifest.Dependencies)
				if err != nil {
					return NewInstallError(err)
				}
				if err := state.AddPackage(info, packages.Oko()...); err != nil {
					return NewInstallError(err)
				}
			} else {
				if err := state.AddPackage(info); err != nil {
					return nil
				}
			}
			if err := state.Save("./oko.json"); err != nil {
				return NewInstallError(err)
			}
			return nil
		}

		// OKO
		if raw, err := os.ReadFile(fmt.Sprintf("%s/oko.json", info.RelativePath())); err == nil {
			pkg, err := config.NewPackageConfig(raw)
			if err != nil {
				return NewInstallError(err)
			}
			if err := state.LoadState(config.NewPackageState(pkg)); err != nil {
				return NewInstallError(err)
			}
			if err := state.Save("./oko.json"); err != nil {
				return NewInstallError(err)
			}
			return nil
		}

		// No `vessel.dhall` or `oko.json`.
		if _, err := os.Stat(fmt.Sprintf("%s/src", info.RelativePath())); err != nil {
			return NewInstallError(err)
		}
		if err := state.AddPackage(info); err != nil {
			return NewInstallError(err)
		}
		if err := state.Save("./oko.json"); err != nil {
			return NewInstallError(err)
		}
		return nil
	},
}

var InstallLocalCommand = cmd.Command{
	Name:        "local",
	Aliases:     []string{"l"},
	Summary:     "install local packages",
	Description: `Allows you to link local packages as dependencies.`,
	Args:        []string{"path"},
	Options: []cmd.Option{
		{
			Name:     "name",
			Summary:  "package name",
			HasValue: true,
		},
	},
	Method: func(args []string, options map[string]string) error {
		path := args[0]
		if _, err := os.Stat(path); err != nil {
			return NewInstallError(err)
		}
		info := config.PackageInfoLocal{
			Path: path,
		}

		name, ok := options["name"]
		if !ok {
			// Ask for rename of package.
			name := path[strings.LastIndex(path, "/")+1:]
			if cmd.AskForConfirmation(fmt.Sprintf("Do you want to rename the package name %q?", name)) {
				info.Name = strings.TrimSpace(cmd.Ask("New name"))
			} else {
				info.Name = name
			}
		} else {
			info.Name = name
		}

		state, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return NewInstallError(err)
		}
		if err := state.AddLocalPackage(info); err != nil {
			return NewInstallError(err)
		}
		if err := state.Save("./oko.json"); err != nil {
			return NewInstallError(err)
		}
		return nil
	},
}

type InstallError struct {
	Err error
}

func NewInstallError(err error) *InstallError {
	return &InstallError{
		Err: err,
	}
}

func (e InstallError) Error() string {
	return fmt.Sprintf("install error %s", e.Err)
}
