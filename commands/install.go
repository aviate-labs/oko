package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/internet-computer/oko/config"
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
		"Expects `{org}/{repo}`, i.e. if you want to install the package at https://github.com/internet-computer/testing.mo you will have to pass `internet-computer/testing.mo` to the first argument.",
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
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}

		if err := info.Download(); err != nil {
			return err
		}
		if rawM, err := os.ReadFile(fmt.Sprintf("%s/vessel.dhall", info.RelativePath())); err == nil {
			manifest, err := vessel.NewManifest(rawM)
			if err != nil {
				return err
			}
			info.Dependencies = manifest.Dependencies
			if len(manifest.Dependencies) != 0 {
				packageSet, err := vessel.LoadPackageSet(fmt.Sprintf("%s/package-set.dhall", info.RelativePath()))
				if err != nil {
					return err
				}
				packages, err := packageSet.Filter(manifest.Dependencies)
				if err != nil {
					return err
				}
				state.AddPackage(info, packages.Oko()...)
			}
		} else {
			if false {
				// Check for Oko packages?
			} else {
				if _, err := os.Stat(fmt.Sprintf("%s/src", info.RelativePath())); err != nil {
					fmt.Println("Invalid packages, no src directory found.")
					return nil
				}
			}
			state.AddPackage(info)
		}
		return state.Save("./oko.json")
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
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("could not find  %q", path)
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
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}
		if err := state.AddLocalPackage(info); err != nil {
			return err
		}
		return state.Save("./oko.json")
	},
}
