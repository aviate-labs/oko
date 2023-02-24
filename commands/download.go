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
			return NewDownloadError(err)
		}
		if err := state.Download(); err != nil {
			return NewDownloadError(err)
		}
		return nil
	},
}

type DownloadError struct {
	Err error
}

func NewDownloadError(err error) *DownloadError {
	return &DownloadError{
		Err: err,
	}
}

func (e DownloadError) Error() string {
	return fmt.Sprintf("download error: %s", e.Err)
}
