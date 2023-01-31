package main

import (
	"fmt"
	"os"

	"github.com/internet-computer/oko/commands"
	"github.com/internet-computer/oko/internal/cmd"
)

const VERSION = "v0.0.0"

var Oko = cmd.Command{
	Name:    "oko",
	Summary: "A Package Manager",
	Commands: append(
		[]cmd.Command{
			VersionCommand,
		},
		commands.Commands...,
	),
}

var VersionCommand = cmd.Command{
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
		Oko.Help()
		return
	}
	if err := Oko.Call(os.Args[1:]...); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
