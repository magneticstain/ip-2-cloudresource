package main

import (
	"log"

	"github.com/magneticstain/ip-2-cloudresource/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
