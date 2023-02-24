package testrunner

import (
	"fmt"
	"testing"
)

func TestDebug(t *testing.T) {
	b, err := justificationToken("http")
	if err != nil {
		t.Fatalf("error is %v", err)
	}
	fmt.Println(string(b))
}
