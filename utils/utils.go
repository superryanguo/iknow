package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func CheckFileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func Runshell(shell string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", shell)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}
func TokenCreate() string {
	ct := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(ct, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token
}
func CheckFileName(prefix string) bool {
	//TODO:
	return true
}
