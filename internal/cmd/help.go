package cmd

import (
	"fmt"
	"strings"
)

func (c Command) Help() {
	if len(c.Description) == 0 {
		fmt.Println(c.Summary)
	} else {
		fmt.Println(strings.TrimSpace(c.Description))
	}
	fmt.Println()
	fmt.Printf("Usage:\n\t%s", c.Name)
	if len(c.Commands) == 0 {
		var args []string
		for _, a := range c.Args {
			args = append(args, fmt.Sprintf("<%s>", a))
		}
		if len(args) != 0 {
			fmt.Printf(" %s", strings.Join(args, " "))
		}
		fmt.Println()

		var options [][]string
		for _, o := range c.Options {
			var option = []string{o.Name}
			if o.HasValue {
				option = append(option, "<value>")
			}
			if len(o.Summary) != 0 {
				option = append(option, o.Summary)
			}
			options = append(options, option)
		}
		if len(options) != 0 {
			fmt.Println()
			fmt.Println("Optional arguments:")
			fmt.Println(FormatTable(options, "\t", "\n", "\t"))
		}
		fmt.Println()
	} else {
		var cmds [][]string
		for _, c := range c.Commands {
			cmds = append(
				cmds,
				[]string{
					fmt.Sprintf("<%s>", c.Name),
					c.Summary,
				},
			)
		}

		fmt.Println(" <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println(FormatTable(cmds, "\t", "\n", "\t"))
	}
}
