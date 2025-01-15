package main

import (
	"log"

	"github.com/primlock/zoomies/cmd"
)

func main() {
	zoomies := cmd.NewCmd()
	err := zoomies.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
