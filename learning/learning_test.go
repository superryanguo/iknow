package learning

import (
	"fmt"
	"testing"
)

func TestPopulateResultToX(t *testing.T) {
	s := "1:0.1 2:0.3 3:12345"
	x, err := PopulateResultToX(s)

	fmt.Println("x=", x)
	if err != nil {
		t.Error("TestPopulateResultToX:", err)
		return
	}

	if x[1] != 0.1 {
		t.Error("wrong value!")
		return
	}
}
