package example

import "testing"

func TestReverse(t *testing.T) {
	res := Reverse("hello")
	if res != "olleh" {
		t.Fatalf("Error, expected olleh but got %s", res)
	}
}
