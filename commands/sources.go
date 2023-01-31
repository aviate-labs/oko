package commands

import (
	"fmt"
	"strings"

	"github.com/internet-computer/oko/config"
	"github.com/internet-computer/oko/internal/cmd"
)

var SourcesCommand = cmd.Command{
	Name:    "sources",
	Summary: "prints moc package sources",
	Method: func(_ []string, _ map[string]string) error {
		state, err := config.LoadPackageState("./oko.json")
		if err != nil {
			return fmt.Errorf("could not load `oko.json`: %s", err)
		}

		var sources []string
		for _, dep := range state.Dependencies {
			sources = append(sources, fmt.Sprintf(
				"--package %s %s/src", dep.Name, dep.RelativePath(),
			))
			for _, name := range dep.AlternativeNames {
				sources = append(sources, fmt.Sprintf(
					"--package %s %s/src", name, dep.RelativePath(),
				))
			}
		}
		for _, dep := range state.TransitiveDependencies {
			sources = append(sources, fmt.Sprintf(
				"--package %s %s/src", dep.Name, dep.RelativePath(),
			))
			for _, name := range dep.AlternativeNames {
				sources = append(sources, fmt.Sprintf(
					"--package %s %s/src", name, dep.RelativePath(),
				))
			}
		}
		for _, dep := range state.LocalDependencies {
			sources = append(sources, fmt.Sprintf(
				"--package %s %s", dep.Name, dep.Path,
			))
		}
		fmt.Print(strings.Join(sources, " "))
		return nil
	},
}
