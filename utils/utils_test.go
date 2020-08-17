package utils

import (
	"fmt"
	"testing"
)

func TestCountFileNum(t *testing.T) {
	dir := "."
	i, err := CountFileNum(dir)
	if err != nil {
		t.Error("TestCountFileNum:", err)
		return
	}
	fmt.Println("FileNum:", i)
	if i != 2 {
		t.Error("Wrong number of files in path")
	}
}
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
