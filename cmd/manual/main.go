package main

import (
	"fmt"

	"github.com/internet-computer/oko/commands"
	"github.com/internet-computer/oko/internal/cmd"
)

func main() {
	fmt.Println(cmd.Manual(commands.Commands))
}
