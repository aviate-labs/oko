package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
	"github.com/internet-computer/oko/vessel"
)

const VERSION = "v0.0.0"

var downloadCommand = cmd.Command{
	Name:    "download",
	Summary: "download packages",
	Method: func(_ []string, _ map[string]string) error {
		pkg, err := config.LoadPackage("./oko.json")
		if err != nil {
			fmt.Printf("Could not load `oko.json`: %s\n", err)
			return nil
		}
		if err := pkg.Download(); err != nil {
			fmt.Printf("Could not download packages: %s\n", err)
		}
		return nil
	},
}

var initCommand = cmd.Command{
	Name: "init",
	Method: func(_ []string, _ map[string]string) error {
		if _, err := config.LoadPackage("./oko.json"); err == nil {
			fmt.Println("`oko.json` already exists.")
			return nil
		}
		if dataM, err := json.MarshalIndent(config.New(), "", "\t"); err == nil {
			return os.WriteFile("oko.json", dataM, os.ModePerm)
		} else {
			return err
		}
	},
}

var installCommand = cmd.Command{
	Name:    "install",
	Summary: "install GitHub hosted packages",
	Args:    cmd.Arguments{"url", "version"},
	Method: func(args []string, _ map[string]string) error {
		url := args[0]
		if !strings.HasPrefix(url, "github.com") {
			fmt.Println("Url needs to start with `github.com`.")
			return nil
		}
		info := config.PackageInfo{
			Repository: fmt.Sprintf("https://%s", url),
			Version:    args[1],
		}
		if name := url[strings.LastIndex(url, "/")+1:]; cmd.AskForConfirmation(fmt.Sprintf("Do you want to rename the package name %q?", name)) {
			info.Name = strings.TrimSpace(cmd.Ask("New name"))
		} else {
			info.Name = name
		}

		raw, err := os.ReadFile("./oko.json")
		if err != nil {
			fmt.Println("Could not find `oko.json` in the current working directory.")
			return nil
		}
		var pkg config.Package
		if err := json.Unmarshal(raw, &pkg); err != nil {
			fmt.Println("Invalid `oko.json` format.")
			return nil
		}
		if _, ok := pkg.Contains(info); ok {
			fmt.Println("Already added to `oko.json`.")
			return nil
		}
		if err := info.Download(); err != nil {
			return err
		}

		rawM, err := os.ReadFile(fmt.Sprintf("%s/vessel.dhall", info.RelativePathDownload()))
		if err == nil {
			manifest, err := vessel.NewManifest(rawM)
			if err != nil {
				return err
			}
			info.Dependencies = manifest.Dependencies

			var replaced = make(map[string]string) // list of replaced dep names
			var newPackages []config.PackageInfo
			if len(manifest.Dependencies) != 0 {
				rawS, err := os.ReadFile(fmt.Sprintf("%s/package-set.dhall", info.RelativePathDownload()))
				if err != nil {
					return err
				}
				packageSet, err := vessel.NewPackageSet(rawS)
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
			pkg.Dependencies = append(pkg.Dependencies, newPackages...)
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
		pkg.Dependencies = append(pkg.Dependencies, info)

		sort.Slice(pkg.Dependencies, func(i, j int) bool {
			return strings.Compare(pkg.Dependencies[i].Name, pkg.Dependencies[j].Name) == -1
		})

		dataM, err := json.MarshalIndent(pkg, "", "\t")
		if err != nil {
			return err
		}
		return os.WriteFile("oko.json", dataM, os.ModePerm)
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
		rawM, err := os.ReadFile("./vessel.dhall")
		if err != nil {
			fmt.Println("Could not find `vessel.dhall` in the current working directory.")
			return nil
		}
		manifest, err := vessel.NewManifest(rawM)
		if err != nil {
			fmt.Println("Could not read `vessel.dhall`.")
			return nil
		}
		rawS, err := os.ReadFile("./package-set.dhall")
		if err != nil {
			fmt.Println("Could not find `package-set.dhall` in the current working directory.")
			return nil
		}
		packageSet, err := vessel.NewPackageSet(rawS)
		if err != nil {
			fmt.Println("Could not read `package-set.dhall`.")
			return nil
		}
		packages, err := packageSet.Filter(manifest.Dependencies)
		if err != nil {
			fmt.Println("Package set incomplete!")
			return nil
		}
		dataM, err := json.MarshalIndent(manifest.Oko(packages), "", "\t")
		if err != nil {
			return err
		}
		if err := os.WriteFile("oko.json", dataM, os.ModePerm); err != nil {
			return err
		}

		// Optional delete.
		if _, ok := options["keep"]; ok {
			return nil
		}
		if _, ok := options["delete"]; ok || cmd.AskForConfirmation("Do you want to delete the `vessel.dhall` file?") {
			if err := os.Remove("./vessel.dhall"); err != nil {
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
		var sources []string
		raw, err := os.ReadFile("./oko.json")
		if err != nil {
			fmt.Println("Could not find `oko.json` in the current working directory.")
			return nil
		}
		var pkg config.Package
		if err := json.Unmarshal(raw, &pkg); err != nil {
			fmt.Println("Invalid `oko.json` format.")
			return nil
		}
		for _, dep := range pkg.Dependencies {
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
		fmt.Printf("\nERR: %s\n\n", err)
		oko.Help()
	}
}
