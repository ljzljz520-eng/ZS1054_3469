package main

import "testing"

func TestCommandPackageCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
}
