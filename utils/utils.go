package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	HOSuccess = "切换成功"
	HOFailure = "切换失败"
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
			files = append(files, path+"/"+file.Name())
		}
	}

	return files, nil
}

func Shuffle(a []string) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

//TODO: how to ignore the ignorefiles, such .swp
func CountFileNum(path string) (int, error) {
	i := 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}
	for _, file := range files {
		if !file.IsDir() {
			i++
		}
	}
	return i, nil
}

func MapMlResult2String(l float64, model string) string {
	switch l {
	case 1:
		if model == "HoSrc" || model == "HoTgt" {
			return HOSuccess
		}
	case -1:
		if model == "HoSrc" || model == "HoTgt" {
			return HOFailure
		}
	}
	return "Unknown string in MapMlResult2String"
}

//TODO: if string is too long?!
func PureDuplicateString(s []string) []string {
	var sn []string

	mapit := make(map[string]bool)

	for _, v := range s {
		if _, ok := mapit[v]; !ok {
			mapit[v] = true
			sn = append(sn, v)
		}
	}
	return sn
}

func DecideEmptyStringHtml(s string) bool {
	if len(strings.Replace(strings.Replace(strings.Replace(strings.Replace(s, "\n", "", -1), "\r", "", -1), " ", "", -1), ";", "", -1)) == 0 {
		return true
	}
	return false
}
