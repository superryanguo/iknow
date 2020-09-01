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
		if v == dir+"/"+"utils.go" {
			flag = true
		}
		fmt.Println("File_", k, ":", v)
	}
	if !flag {
		t.Error("Miss the uitls.go")
	}

}
func TestPureDuplicateString(t *testing.T) {
	st := []string{"abc", "sfds", "abc"}
	target := []string{"abc", "sfds"}
	si := PureDuplicateString(st)
	fmt.Println("New slice after del duplicate:", si)
	if len(target) != len(si) {
		t.Error("PureDuplicateString error")
	}
}
func TestDecideEmptyStringHtml(t *testing.T) {
	st := "      \n\r;"
	si := DecideEmptyStringHtml(st)
	if !si {
		t.Error("DecideEmptyStringHtml error")
	}
}
