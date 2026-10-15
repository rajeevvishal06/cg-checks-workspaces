package main

import "testing"

func TestGreet(t *testing.T) {
	if Greet() == "" {
		t.Fatal("expected a non-empty greeting")
	}
}
