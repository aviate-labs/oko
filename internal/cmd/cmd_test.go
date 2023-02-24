package cmd_test

import (
	"fmt"

	"github.com/internet-computer/oko/internal/cmd"
)

var (
	s = cmd.Command{
		Name: "sub",
		Args: []string{"c"},
		Options: []cmd.Option{
			{"all", "", false},
			{"v", "", true},
		},
		Method: func(args []string, options map[string]string) error {
			fmt.Println(args, options)
			return nil
		},
	}
	c = cmd.Command{
		Name:     "test",
		Aliases:  []string{"t"},
		Commands: []cmd.Command{s},
	}
)

func ExampleCommand_Help() {
	c.Help()
	// Output:
	// Usage:
	// 	test <command>
	//
	// Commands:
	// 	<sub>
}

func ExampleCommand_Help_sub() {
	_ = c.Call("sub", "help")
	// Output:
	// Usage:
	//	sub <c>
	//
	// Optional arguments:
	//	all
	//	v  	<value>
}

func ExampleCommand_Help_subC() {
	_ = c.Call("sub", "c")
	_ = c.Call("sub", "--all", "c")
	_ = c.Call("sub", "c", "--all")
	_ = c.Call("sub", "c", "--v=0")
	_ = c.Call("sub", "--v", "0", "c")
	// Output:
	// [c] map[]
	// [c] map[all:]
	// [c] map[all:]
	// [c] map[v:0]
	// [c] map[v:0]
}
