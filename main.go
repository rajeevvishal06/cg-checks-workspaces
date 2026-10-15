package main

import "fmt"

// Greet returns the demo banner.
func Greet() string {
	return "Chainguard Checks + Workspaces demo"
}

func main() {
	fmt.Println(Greet())
}
