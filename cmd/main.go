package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
	"github.com/internet-computer/oko/vessel"
)

const VERSION = "v0.0.0"

var downloadCommand = cmd.Command{
	Name:    "download",
	Aliases: []string{"d"},
	Summary: "download packages",
	Method: func(_ []string, _ map[string]string) error {
		pkg, err := config.LoadPackage("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}
		if err := pkg.Download(); err != nil {
			return fmt.Errorf("could not download packages: %s", err)
		}
		return nil
	},
}

var initCommand = cmd.Command{
	Name:    "init",
	Summary: "initialize Oko",
	Method: func(_ []string, _ map[string]string) error {
		if _, err := config.LoadPackage("./oko.json"); err == nil {
			return fmt.Errorf("`oko.json` already exists: %s", err)
		}
		return config.EmptyPackage().Save("./oko.json")
	},
}

var installCommand = cmd.Command{
	Name:    "install",
	Aliases: []string{"i"},
	Summary: "install packages",
	Commands: []cmd.Command{
		installGitHubCommand,
		installLocalCommand,
	},
}

var installLocalCommand = cmd.Command{
	Name:    "local",
	Aliases: []string{"l"},
	Summary: "install local packages",
	Args:    []string{"path"},
	Options: []cmd.Option{
		{
			Name:     "name",
			Summary:  "local path",
			HasValue: true,
		},
	},
	Method: func(args []string, options map[string]string) error {
		path := args[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("could not find  %q", path)
		}
		local := true
		info := config.PackageInfo{
			Repository: path,
			Local:      &local,
			Version:    "*",
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

		pkg, err := config.LoadPackage("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}
		if _, ok := pkg.Contains(info); ok {
			return fmt.Errorf("already added to `oko.json`")
		}
		pkg.Add(info)
		return pkg.Save("./oko.json")
	},
}

var installGitHubCommand = cmd.Command{
	Name:    "github",
	Aliases: []string{"gh"},
	Summary: "install GitHub hosted packages",
	Args:    []string{"url", "version"},
	Method: func(args []string, _ map[string]string) error {
		url := args[0]
		if !strings.HasPrefix(url, "github.com") {
			return fmt.Errorf("url needs to start with `github.com`")
		}
		version := args[1]
		info := config.PackageInfo{
			Repository: fmt.Sprintf("https://%s", url),
			Version:    version,
		}

		// Ask for rename of package.
		if name := url[strings.LastIndex(url, "/")+1:]; cmd.AskForConfirmation(fmt.Sprintf("Do you want to rename the package name %q?", name)) {
			info.Name = strings.TrimSpace(cmd.Ask("New name"))
		} else {
			info.Name = name
		}

		pkg, err := config.LoadPackage("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}
		if _, ok := pkg.Contains(info); ok {
			return fmt.Errorf("already added to `oko.json`")
		}
		if err := info.Download(); err != nil {
			return err
		}

		if rawM, err := os.ReadFile(fmt.Sprintf("%s/vessel.dhall", info.RelativePathDownload())); err == nil {
			manifest, err := vessel.NewManifest(rawM)
			if err != nil {
				return err
			}
			info.Dependencies = manifest.Dependencies

			// List of replaced dependency names.
			var replaced = make(map[string]string)
			var newPackages []config.PackageInfo
			if len(manifest.Dependencies) != 0 {
				packageSet, err := vessel.LoadPackageSet(fmt.Sprintf("%s/package-set.dhall", info.RelativePathDownload()))
				if err != nil {
					return err
				}
				packages, err := packageSet.Filter(manifest.Dependencies)
				if err != nil {
					return err
				}
				for i, dep := range packages.Oko() {
					if err := dep.Download(); err != nil {
						return err
					}
					if name, exists := pkg.Contains(dep); !exists {
						newPackages = append(newPackages, dep)
					} else {
						replaced[manifest.Dependencies[i]] = name
						manifest.Dependencies[i] = name
					}
				}
			}
			for _, dep := range newPackages {
				for i, d := range dep.Dependencies {
					if n, ok := replaced[d]; ok {
						dep.Dependencies[i] = n
					}
				}
			}
			pkg.Add(newPackages...)
		} else {
			if false {
				// Check for Oko packages?
			} else {
				if _, err := os.Stat(fmt.Sprintf("%s/src", info.RelativePathDownload())); err != nil {
					fmt.Println("Invalid packages, no src directory found.")
					return nil
				}
			}
		}

		pkg.Add(info)
		return pkg.Save("./oko.json")
	},
}

var migrateCommand = cmd.Command{
	Name:    "migrate",
	Summary: "migrate Vessel packages",
	Options: []cmd.Option{
		{
			Name:     "delete",
			HasValue: false,
		},
		{
			Name:     "keep",
			HasValue: false,
		},
	},
	Method: func(_ []string, options map[string]string) error {
		if 2 <= len(options) {
			return fmt.Errorf("can not use both `delete` and `keep` at the same time")
		}

		manifest, err := vessel.LoadManifest("./vessel.dhall")
		if err != nil {
			return fmt.Errorf("could not read `vessel.dhall`: %s", err)
		}
		packageSet, err := vessel.LoadPackageSet("./package-set.dhall")
		if err != nil {
			return fmt.Errorf("could not read `package-set.dhall`: %s", err)
		}

		packages, err := packageSet.Filter(manifest.Dependencies)
		if err != nil {
			return fmt.Errorf("package set incomplete: %s", err)
		}

		if err := manifest.Save("./oko.json", packages); err != nil {
			return fmt.Errorf("failed saving package set: %s", err)
		}

		// Optional delete.
		if _, ok := options["keep"]; ok {
			return nil
		}
		if _, ok := options["delete"]; ok || cmd.AskForConfirmation("Do you want to delete the `vessel.dhall` and `package-set.dhall` file?") {
			if err := os.Remove("./vessel.dhall"); err != nil {
				return err
			}
			if err := os.Remove("./package-set.dhall"); err != nil {
				return err
			}
		}
		return nil
	},
}

var oko = cmd.Command{
	Name:    "oko",
	Summary: "TODO: fix me",
	Commands: []cmd.Command{
		versionCommand,
		initCommand,
		downloadCommand,
		installCommand,
		migrateCommand,
		sourcesCommand,
	},
}

var sourcesCommand = cmd.Command{
	Name:    "sources",
	Summary: "prints moc package sources",
	Method: func(_ []string, _ map[string]string) error {
		pkg, err := config.LoadPackage("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}

		var sources []string
		for _, dep := range pkg.Dependencies {
			if dep.IsLocal() {
				sources = append(sources, fmt.Sprintf(
					"--package %s %s", dep.Name, dep.RelativePathDownload(),
				))
				continue
			}

			sources = append(sources, fmt.Sprintf(
				"--package %s %s/src", dep.Name, dep.RelativePathDownload(),
			))
			for _, name := range dep.AlternativeNames {
				sources = append(sources, fmt.Sprintf(
					"--package %s %s/src", name, dep.RelativePathDownload(),
				))
			}
		}
		fmt.Print(strings.Join(sources, " "))
		return nil
	},
}

var versionCommand = cmd.Command{
	Name:    "version",
	Aliases: []string{"v"},
	Summary: "print Oko version",
	Method: func(args []string, _ map[string]string) error {
		fmt.Println(VERSION)
		return nil
	},
}

func main() {
	if len(os.Args) == 1 {
		oko.Help()
		return
	}
	if err := oko.Call(os.Args[1:]...); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
