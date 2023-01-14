package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func Ask(q string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", q)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return response
}

func AskForConfirmation(q string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [Y/n]: ", q)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "" || response == "y" || response == "ye" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
