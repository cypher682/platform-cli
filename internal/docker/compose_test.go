package docker

import (
	"reflect"
	"testing"
)

func TestComposeArgs(t *testing.T) {
	got := Compose{}.args("up", "-d")
	if !reflect.DeepEqual(got, []string{"compose", "up", "-d"}) {
		t.Fatalf("got %v", got)
	}
	got = Compose{File: "dc.yml"}.args("down", "-v")
	want := []string{"compose", "-f", "dc.yml", "down", "-v"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
