package commands

import (
	"fmt"
	"os"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
	"github.com/internet-computer/oko/vessel"
)

var MigrateCommand = cmd.Command{
	Name:        "migrate",
	Summary:     "migrate Vessel packages",
	Description: `Allows you to migrate Vessel config files to Oko.`,
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
		if _, err := config.LoadPackageState("./oko.json"); err == nil {
			return fmt.Errorf("can not migrate vessel packages, `oko.json` already exists")
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
