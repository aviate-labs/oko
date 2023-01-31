package commands

import (
	"fmt"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
)

var RemoveCommand = cmd.Command{
	Name:        "remove",
	Aliases:     []string{"r"},
	Summary:     "remove a package",
	Description: `Allows you to remove packages by name.`,
	Args:        []string{"name"},
	Method: func(args []string, _ map[string]string) error {
		state, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}

		name := args[0]
		if err := state.RemoveLocalPackage(name); err == nil {
			return state.Save("./oko.json")
		}

		if err := state.RemovePackage(name); err != nil {
			return fmt.Errorf("could not remove package: %s", err)
		}
		return state.Save("./oko.json")
	},
}
