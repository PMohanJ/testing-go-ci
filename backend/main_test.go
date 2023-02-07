package main

import (
	"testing"
)

func TestTemp(t *testing.T) {
	t.Run("prints some text", func(t *testing.T) {
		want := "Some temporary function"

		got := Temp()

		if got != want {
			t.Errorf("Got %v, want %v", got, want)
		}
	})
}
