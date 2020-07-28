package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func FilterFileList(path, suffix string) ([]string, error) {
	var files []string

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range fileInfo {
		if strings.HasSuffix(file.Name(), suffix) {
			files = append(files, file.Name())
		}
	}

	return files, nil
}
