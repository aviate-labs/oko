package cmd

import (
	"fmt"
	"strings"
)

func trimPrefix(s, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix), true
	}
	return s, false
}

// A command with either sub-commands or a list of arguments.
type Command struct {
	// The name of the command.
	Name string
	// Aliases of the name.
	// e.g. version -> v
	Aliases []string
	// A summary explaining the function of the command.
	Summary string
	// A longer description of the command.
	Description string

	// A list of sub commands.
	Commands []Command

	// A list of arguments.
	Args []string
	// Options of the command.
	// e.g. --all, etc.
	Options []Option
	// The method corresponding with a list of arguments.
	Method func(args []string, options map[string]string) error
}

func (c Command) Call(args ...string) error {
	if c.Method != nil {
		return c.method(args)
	}
	if len(args) == 0 || args[0] == "help" {
		c.Help()
		return nil
	}
	return c.command(args[0], args[1:])
}

// checkArguments returns an error if the number of arguments do not equal the
// expected amount.
func (c Command) checkArguments(args []string) error {
	l := len(c.Args)
	if len(args) != l {
		var s []string
		for _, a := range c.Args {
			s = append(s, fmt.Sprintf("<%s>", a))
		}

		switch l {
		case 0:
			return NewInvalidArgumentsError("expected no argument")
		case 1:
			return NewInvalidArgumentsError(fmt.Sprintf("expected 1 argument: %s", s[0]))
		default:
			return NewInvalidArgumentsError(fmt.Sprintf("expected %d argument(s): %s", len(c.Args), strings.Join(s, " ")))
		}
	}
	return nil
}

func (c Command) command(name string, args []string) error {
	var cmd Command
	for _, c := range c.Commands {
		for _, n := range append([]string{c.Name}, c.Aliases...) {
			if n == name {
				cmd = c
				break
			}
		}
	}
	if cmd.Name == "" {
		return NewCommandNotFoundError()
	}
	return cmd.Call(args...)
}

func (c Command) extractOptions(args []string) ([]string, map[string]string) {
	var (
		arguments []string
		arg       string
		options   = make(map[string]string)
	)
	for _, a := range args {
		if arg != "" {
			options[arg] = a
			arg = ""
			continue
		}

		if a, ok := trimPrefix(a, "--"); ok {
			var cont bool
			for _, o := range c.Options {
				if a, ok := trimPrefix(a, o.Name); ok {
					if a, ok := trimPrefix(a, "="); ok && o.HasValue {
						if a != "" {
							options[o.Name] = a
							cont = true
							break
						}
					}
					if a == "" {
						if o.HasValue {
							arg = o.Name
						} else {
							options[o.Name] = ""
						}
						cont = true
						break
					}
				}
			}
			if cont {
				continue
			}
		}
		arguments = append(arguments, a)
	}
	return arguments, options
}

func (c Command) method(args []string) error {
	if len(args) == 1 && args[0] == "help" {
		c.Help()
		return nil
	}
	args, opts := c.extractOptions(args)
	if err := c.checkArguments(args); err != nil {
		c.Help()
		return nil
	}
	if err := c.Method(args, opts); err != nil {
		return err
	}
	return nil
}

type Option struct {
	Name     string
	Summary  string
	HasValue bool
}
