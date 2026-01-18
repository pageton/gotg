package main

import "fmt"

// TODO: parse types and try to remove hardcodedreplacements

func main() {
	fmt.Println("GoTG Files Generator (c)2023")
	fmt.Println("Running Generator...")
	fmt.Println("Generating generic helpers for context.go")
	generateCUHelpers()
	fmt.Println("Generated Successfully!")
}
