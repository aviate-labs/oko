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
			return NewRemoveError(err)
		}

		name := args[0]
		if err := state.RemoveLocalPackage(name); err == nil {
			if err := state.Save("./oko.json"); err != nil {
				return NewRemoveError(err)
			}
			return nil
		}

		if err := state.RemovePackage(name); err != nil {
			return NewRemoveError(err)
		}
		if err := state.Save("./oko.json"); err != nil {
			return NewRemoveError(err)
		}
		return nil
	},
}

type RemoveError struct {
	Err error
}

func NewRemoveError(err error) *RemoveError {
	return &RemoveError{
		Err: err,
	}
}

func (e RemoveError) Error() string {
	return fmt.Sprintf("remove error: %s", e.Err)
}
