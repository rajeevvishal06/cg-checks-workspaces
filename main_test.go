package main

import "testing"

func TestGreet(t *testing.T) {
	// Deliberately wrong expectation so `go test` fails — to see the gate catch it.
	if Greet() != "this will never match" {
		t.Fatalf("intentional failure: Greet() = %q", Greet())
	}
}
