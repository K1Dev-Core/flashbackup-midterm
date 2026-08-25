package main

import "testing"

func TestCompleteCommand(t *testing.T) {
	got, matches := completeCommand("/sou")
	if got != "/source " || matches != nil {
		t.Fatalf("completeCommand(/sou) = %q, %v", got, matches)
	}
	got, matches = completeCommand("/source ")
	if got != "/source " || matches != nil {
		t.Fatalf("completeCommand(/source ) = %q, %v", got, matches)
	}
	got, matches = completeCommand("/")
	if got != "/" || len(matches) != len(commands) {
		t.Fatalf("completeCommand(/) = %q, %v", got, matches)
	}
}
