package main

import "testing"

func TestCommandMode(t *testing.T) {
	if got, err := commandMode(nil); err != nil || got != modeWorker {
		t.Fatalf("no arguments = %v, %v; want the worker", got, err)
	}
	if got, err := commandMode([]string{"regenerate-masters"}); err != nil || got != modeRegenerateMasters {
		t.Fatalf("regenerate-masters = %v, %v", got, err)
	}
	for _, arguments := range [][]string{{"regenerate"}, {"regenerate-masters", "now"}, {"--help"}} {
		if _, err := commandMode(arguments); err == nil {
			t.Fatalf("%v accepted", arguments)
		}
	}
}
