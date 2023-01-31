package commands

import (
	"fmt"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
)

var InitCommand = cmd.Command{
	Name:        "init",
	Summary:     "initialize Oko",
	Description: `Initializes the Oko package file.`,
	Options: []cmd.Option{
		{
			Name:     "compiler",
			Summary:  "compiler version",
			HasValue: true,
		},
	},
	Method: func(_ []string, options map[string]string) error {
		if _, err := config.LoadPackageState("./oko.json"); err == nil {
			return fmt.Errorf("`oko.json` already exists: %s", err)
		}
		state := config.EmptyState()
		if v, ok := options["compiler"]; ok {
			state.CompilerVersion = &v
		}
		return state.Save("./oko.json")
	},
}
