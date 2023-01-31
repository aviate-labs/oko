package commands

import "github.com/internet-computer/oko/internal/cmd"

var Commands = []cmd.Command{
	InitCommand,
	DownloadCommand,
	InstallCommand,
	RemoveCommand,
	MigrateCommand,
	SourcesCommand,
	BinCommand,
}
