package cmd

import (
	"fmt"
	"strings"
)

// Returns a MD styled docs for the given list of commands.
func Manual(commands []Command) string {
	return strings.TrimSpace(manual("Commands", 1, commands))
}

func headerPrefix(indent int) string {
	return strings.Repeat("#", indent)
}

func manual(prefix string, indent int, commands []Command) string {
	var man string = fmt.Sprintf("%s %s\n\n", headerPrefix(indent), prefix)
	for _, cmd := range commands {
		// Name and description.
		desc := strings.TrimSpace(cmd.Description)
		if len(desc) == 0 {
			desc = cmd.Summary
		}
		man += fmt.Sprintf(
			"%s `%s`\n\n%s\n\n",
			headerPrefix(indent+1), cmd.Name, desc,
		)

		// Sub-commands
		if len(cmd.Commands) != 0 {
			man += manual("Sub Commands", indent+2, cmd.Commands)
		} else {
			if len(cmd.Args) != 0 {
				man += fmt.Sprintf("%s Arguments\n\n", headerPrefix(indent+2))
				for i, arg := range cmd.Args {
					man += fmt.Sprintf("%d. %s\n", i+1, arg)
				}
				man += "\n"
			}
			if len(cmd.Options) != 0 {
				man += fmt.Sprintf("%s Options\n\n", headerPrefix(indent+2))
				man += "|name|value|\n|---|---|\n"
				for _, o := range cmd.Options {
					man += fmt.Sprintf("|**%s**|", o.Name)
					if o.HasValue {
						if len(o.Summary) != 0 {
							man += fmt.Sprintf("*%s*", o.Summary)
						} else {
							man += "*value*"
						}
					}
					man += "|\n"
				}
				man += "\n"
			}
		}
	}
	return man
}
