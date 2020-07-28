package utils

import (
	"fmt"
	"testing"
)

func TestFilterFileList(t *testing.T) {
	flag := false
	dir := "."
	fl, err := FilterFileList(dir, ".go")
	if err != nil {
		t.Error("FilterFileList:", err)
		return
	}
	for k, v := range fl {
		if v == "utils.go" {
			flag = true
		}
		fmt.Println("File_", k, ":", v)
	}
	if !flag {
		t.Error("Miss the uitls.go")
	}

}
