package cmd_test

import (
	"fmt"

	"github.com/internet-computer/oko/internal/cmd"
)

func Example_formatTable() {
	fmt.Println(cmd.FormatTable([][]string{
		{"_", "__", "___"},
		{"a", "b", "c"},
		{"____", "d", "_", "e"},
	}, " | ", "\n", "| "))
	// Output:
	// | _    | __ | ___
	// | a    | b  | c
	// | ____ | d  | _   | e
}
