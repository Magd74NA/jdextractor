package main

import (
	"fmt"
	"jdextract/jdextract"
)

func main() {
	fmt.Println("Hello World!")
	jdextract.CheckWiring()
	jdextract.NewApp()
	fmt.Println("New app created!")
}
