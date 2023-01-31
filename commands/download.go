package commands

import (
	"fmt"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
)

var DownloadCommand = cmd.Command{
	Name:        "download",
	Aliases:     []string{"d"},
	Summary:     "download packages",
	Description: `Downloads all packages specified in the Oko package file.`,
	Method: func(_ []string, _ map[string]string) error {
		state, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}
		if err := state.Download(); err != nil {
			return fmt.Errorf("could not download packages: %s", err)
		}
		return nil
	},
}
