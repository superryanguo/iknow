package feature

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/utils"
)

var MsgMap FeatureMsgMap
var MsgTpt FeatureTemplate

const (
	SeqIndex   = 0
	MsgIndex   = 1
	PointIndex = 2
	PageSize   = 1024
	TptSize    = 0
	TptCap     = 10
	regx       = "\\[((([0-9]{3}[1-9]|[0-9]{2}[1-9][0-9]{1}|[0-9]{1}[1-9][0-9]{2}|[1-9][0-9]{3})-(((0[13578]|1[02])-(0[1-9]|[12][0-9]|3[01]))|((0[469]|11)-(0[1-9]|[12][0-9]|30))|(02-(0[1-9]|[1][0-9]|2[0-8]))))|((([0-9]{2})(0[48]|[2468][048]|[13579][26])|((0[48]|[2468][048]|[3579][26])00))-02-29))\\s+([0-1]?[0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])\\.[0-9]*]"
)

func init() {
	MsgMap.M = make(map[string]int)
	MsgTpt.T = make([]Template, TptSize, TptCap)
}

type Template struct {
	Seq     int
	MsgName string
	Point   int
}

type FeatureRaw struct {
	Time  time.Time //Message Time
	Value string    //Message Name
	Flag  int       //TBD: 1 pure 0 not pure
}

type FeaturePure struct {
	Time  int //delta time
	Value int //message value from a map
	Flag  int
}

//A FeatureMsgMap is how the message to map to a int point
type FeatureMsgMap struct {
	M map[string]int
}

type FeatureTemplate struct {
	T []Template
}

func (t *FeatureTemplate) Print() {
	fmt.Println("++++++++FeatureTemplate+++++++++++++")
	fmt.Println("Seq---------->Msg---------->Point")
	for i := 0; i < len(t.T); i++ {
		a := t.T[i]
		fmt.Printf("%d------->%s --------> %d\n", a.Seq, a.MsgName, a.Point)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
}

func (m *FeatureMsgMap) Print() {
	fmt.Println("++++++++FeatureMsgMap+++++++++++++")
	fmt.Println("Msg------->Point")
	for k, v := range m.M {
		fmt.Printf("%s ------> %d\n", k, v)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
}

func CheckDuplicateMsg(t *FeatureTemplate) error {
	//TODO: the same name in the feature template
	//if the same message name, such as Request->Request1/Request2
	return nil
}

func (m *FeatureMsgMap) Build(t FeatureTemplate) error {
	if t.T == nil {
		return errors.New("Empty FeatureTemplate data")
	}

	for i := 0; i < len(t.T); i++ {
		ti := t.T[i]
		m.M[ti.MsgName] = ti.Point
	}
	return nil
}

//ExtractFeatureTemplate extract the feature template from file
//First element should be seq, 2nd be message name, 3rd shoudl be point
//Divided by space
func ExtractFeatureTemplate(file string) (t FeatureTemplate, e error) {
	if !utils.CheckFileExist(file) {
		return FeatureTemplate{nil}, errors.New("File not exist")
	}

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	t.T = make([]Template, TptSize, TptCap)
	s := bufio.NewScanner(f)

	for s.Scan() {
		r := strings.Fields(s.Text())
		//log.Debug("Split each line  to =", r)
		if len(r) != (PointIndex + 1) {
			return FeatureTemplate{nil}, errors.New("Wrong format of tempalte files, not 3 IE")
		}
		j := Template{}
		j.Seq, err = strconv.Atoi(r[SeqIndex])
		if err != nil {
			return FeatureTemplate{nil}, err
		}
		j.MsgName = r[MsgIndex]
		j.Point, err = strconv.Atoi(r[PointIndex])
		if err != nil {
			return FeatureTemplate{nil}, err
		}

		t.T = append(t.T, j)
	}

	return t, nil

}

func CaptureFeautresPlus(file string) ([]string, error) {

	if !utils.CheckFileExist(file) {
		return nil, errors.New("Log file not exist")
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if len(MsgMap.M) == 0 {
		return nil, errors.New("MsgMap is empety, pls init it")
	}

	t := make([]string, TptSize, TptCap)
	buffer := make([]byte, PageSize)

	for {
		lpos, err := f.Seek(0, 1)
		if err != nil {
			return nil, err
		}

		_, err = f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		} else {

			r := strings.Fields(string(buffer))
			log.Debug("Split each line  to =", r)
			if len(r) != 0 {
				for _, v := range r {
					if _, ok := MsgMap.M[v]; ok {
						j := strings.Index(string(buffer), v)

						_, err := f.Seek(lpos+int64(j+len(v)), 0)
						if err != nil {
							return nil, err
						}
						rt, err := SeekTime(f, 500)
						if err != nil || len(rt) == 0 {
							log.Info("Can't find timestamp for->", v)
							t = append(t, v)
							continue
						}
						t = append(t, rt+v)
					}
				}
			}
		}
	}

	return t, nil
}
func CaptureFeautres(file string) ([]string, error) {

	if !utils.CheckFileExist(file) {
		return nil, errors.New("Log file not exist")
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if len(MsgMap.M) == 0 {
		return nil, errors.New("MsgMap is empety, pls init it")
	}

	t := make([]string, TptSize, TptCap)
	s := bufio.NewScanner(f)
	rp := regexp.MustCompile(regx)

	for s.Scan() {
		r := strings.Fields(s.Text())
		//log.Debug("Split each line  to =", r)
		if len(r) != 0 {
			for _, v := range r {
				if _, ok := MsgMap.M[v]; ok {
					sr := rp.FindAllString(s.Text(), -1)
					if sr != nil {
						t = append(t, sr[0]+v)
					} else {
						rt, err := SeekTime(f, 500)
						if err != nil || len(rt) == 0 {
							log.Info("Can't find timestamp for->", v)
							t = append(t, v)
							continue
						}
						t = append(t, rt+v)
					}

				}
			}
		}
	}

	return t, nil
}

//SeekTime search the most close timestamp before current position
func SeekTime(file *os.File, size int64) (string, error) {
	buffer := make([]byte, size)
	cpos, err := file.Seek(0, 1)
	if err != nil {
		return "", err
	}
	log.Debug("FileKeepCurrentOffset=", cpos)

	rp := regexp.MustCompile(regx)
	npos := cpos - size
	var result string

	for npos > 0 {
		log.Debug("npos=", npos, "size=", size)
		_, err := file.ReadAt(buffer, npos)
		if err != nil {
			return "", err
		}
		//log.Debugf("Read %d buffer=>%s", length, buffer)
		sr := rp.FindAllString(string(buffer), -1)
		if sr != nil {
			l := len(sr)
			result = sr[l-1]
			log.Debug("Found closest timestamp", result)
			break
		}
		npos = npos - size
	}

	if npos <= 0 {
		log.Debug("buf npos=", npos, "size=", size)
		buf := make([]byte, size+npos)
		_, err := file.ReadAt(buf, 0)
		if err != nil {
			return "", err
		}
		//log.Debugf("Read %d buf=>%s", length, buf)
		sr := rp.FindAllString(string(buf), -1)
		if sr != nil {
			l := len(sr)
			result = sr[l-1]
			log.Debug("Found closest timestamp", result)
		}
	}

	pos, err := file.Seek(cpos, 0)
	if err != nil {
		return "", err
	}
	log.Debug("FileBackCurrentOffset=", pos)

	if len(result) == 0 {
		return "", errors.New("Can't find the timestamp")
	}

	return result, nil
}
