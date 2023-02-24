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
			return NewMigrateError(NewOptionsError("can not use both `delete` and `keep` at the same time"))
		}
		if _, err := config.LoadPackageState("./oko.json"); err == nil {
			return NewMigrateError(err)
		}

		manifest, err := vessel.LoadManifest("./vessel.dhall")
		if err != nil {
			return NewMigrateError(err)
		}
		packageSet, err := vessel.LoadPackageSet("./package-set.dhall")
		if err != nil {
			return NewMigrateError(err)
		}

		packages, err := packageSet.Filter(manifest.Dependencies)
		if err != nil {
			return NewMigrateError(err)
		}

		if err := manifest.Save("./oko.json", packages); err != nil {
			return NewMigrateError(err)
		}

		// Optional delete.
		if _, ok := options["keep"]; ok {
			return nil
		}
		if _, ok := options["delete"]; ok || cmd.AskForConfirmation("Do you want to delete the `vessel.dhall` and `package-set.dhall` file?") {
			if err := os.Remove("./vessel.dhall"); err != nil {
				return NewMigrateError(err)
			}
			if err := os.Remove("./package-set.dhall"); err != nil {
				return NewMigrateError(err)
			}
		}
		return nil
	},
}

type MigrateError struct {
	Err error
}

func NewMigrateError(err error) *MigrateError {
	return &MigrateError{
		Err: err,
	}
}

func (e MigrateError) Error() string {
	return fmt.Sprintf("migrate error: %s", e.Err)
}
