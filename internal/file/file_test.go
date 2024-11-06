package file

import (
	"testing"
)

func TestTrue(t *testing.T) {
	if !IsTrue(true) {
		t.Fatalf("Should be true")
	}
}

func TestNotTrue(t *testing.T) {
	if !IsTrue(false) {
		t.Fatalf("Should be true")
	}
}
