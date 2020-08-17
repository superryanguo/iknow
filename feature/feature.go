package feature

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/utils"
)

var MsgMap FeatureMsgMap //TODO:if multi-user, will it conflict?! Make it per user?!
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

type TestStatus struct {
	Seq     int
	MsgName string
}

type Template struct {
	Seq     int
	MsgName string
	Point   int
}

type FeatureRawChain []FeatureRaw

type FeatureRaw struct {
	Time  time.Time //Message Time
	Nano  int64
	Value string //Message Name
	Flag  int    //TBD: 1 pure 0 not pure
}

type FeaturePureChain []FeaturePure
type FeaturePure struct {
	Time   time.Time //Message Time
	Nano   int64
	BoTime int64   //delta time from the msg0
	DeTime int64   //delta time from the previous msg
	Value  string  //message value from a map
	NorVal float64 //normalize feature value
	Flag   int
}

//A FeatureMsgMap is how the message to map to a int point
type FeatureMsgMap struct {
	M map[string]int
}

type FeatureTestStatus struct {
	S []TestStatus
}
type FeatureTemplate struct {
	T []Template
}

func (m FeatureTestStatus) Print() {
	fmt.Println("++++++++FeatureTestStatus+++++++++++++")
	fmt.Println("Index------Seq------->Msg")
	for k, v := range m.S {
		fmt.Printf("%d ------> %v\n", k, v)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
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
		fmt.Printf("%s ------> %v\n", k, v)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
}

func (m FeatureRawChain) Print() {
	fmt.Println("++++++++FeatureRaw+++++++++++++")
	fmt.Println("Index-->Time----Nano-----------Msg-------Flag")
	for k, v := range m {
		fmt.Printf("%d ------> %v\n", k, v)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
}

func (m FeaturePureChain) Print() {
	fmt.Println("++++++++FeaturePure+++++++++++++")
	fmt.Println("Index-->Time----Nano-----BoTime---DeTime---Msg-------Flag")
	for k, v := range m {
		fmt.Printf("%d ------> %v\n", k, v)
	}
	fmt.Println("++++++++++++++++++++++++++++++++++")
}

//SvmScaleDeTimeNormalize map the detime to the range 1-2
//TODO: it's better we can handle the negative value to indicate the time before the next message
//Scale the int to float in the featurepure
func (fr *FeaturePureChain) SvmDeTimeNormalize(min, max int64) error {
	delta := float64(max - min)
	if len(*fr) == 0 {
		return errors.New("the FeaturePureChain length is zero!")
	}

	var big, small int64
	for k, v := range *fr {
		if k == 0 {
			big = v.DeTime
			small = v.DeTime
		}

		if big < v.DeTime {
			big = v.DeTime
		}

		if small > v.DeTime {
			small = v.DeTime
		}
	}

	//if 0, should be ajust to 1
	//if samll==0{
	//small=1
	//}

	log.Debug("SvmDeTimeNormalize Small: ", small, " Big: ", big)

	//for _, v := range *fr {
	//v.NorVal = delta*float64(v.DeTime-small)/float64(big-small) + float64(min)
	//log.Debug("SvmDeTimeNormalize NorVal ", v.NorVal)
	//}
	for i := 0; i < len(*fr); i++ {
		(*fr)[i].NorVal = delta*float64((*fr)[i].DeTime-small)/float64(big-small) + float64(min)
	}

	return nil
}

func CheckDuplicateMsg(t FeatureTemplate) error {
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
//First element should be seq(start from 1, not 0), 2nd be message name, 3rd should be point
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
func ExtractFeatureTemplateHtml(input string) (t FeatureTemplate, err error) {
	if len(input) == 0 {
		return FeatureTemplate{nil}, errors.New("Empty Template Input String")
	}
	t.T = make([]Template, TptSize, TptCap)
	datas := strings.Split(input, "\n")

	for _, v := range datas {
		r := strings.Fields(v)
		log.Debug("Split each line  to =", r)
		if len(r) == 0 {
			continue
		}
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
	if len(t.T) == 0 {
		return FeatureTemplate{nil}, errors.New("Empty Template Input")
	}
	return t, nil
}

func BuildTestStatus(r FeaturePureChain) (fs FeatureTestStatus) {
	fs.S = make([]TestStatus, TptSize, TptCap)
	for i := 0; i < len(r); i++ {
		v := TestStatus{}
		v.Seq = i + 1 //begin from 1
		v.MsgName = r[i].Value
		fs.S = append(fs.S, v)
	}
	return
}

//CaptureFeatures make a exact split file with the first timestamp in logs
func CaptureFeatures(file string, full bool) ([]FeatureRaw, error) {
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

	r, err := SeekTimeSlam(f, 100)
	if err != nil {
		return nil, err
	}

	if len(r) == 0 {
		return nil, errors.New("wrong length of the SeekTimeSlam result")
	}

	t := make([]FeatureRaw, TptSize, TptCap)

	for _, w := range r {
		s := strings.Fields(w)
		for _, v := range s {
			if _, ok := MsgMap.M[v]; ok {
				var raw FeatureRaw
				raw.Value = v
				rt, err := PickFirstTime(w)
				if err != nil {
					return nil, err
				}
				wt := strings.Replace(strings.Replace(rt, "[", "", -1), "]", "", -1)
				log.Debug("CaptureFeatures time:", wt, "value=", v)
				raw.Time, raw.Nano, err = FindTimeNano(wt)
				if err != nil {
					return nil, fmt.Errorf("FindTimeNano Error:%s", err)
				}
				t = append(t, raw)

			}
		}
	}

	//if full false, we need the msg0 as the start point
	if !full {
		if len(MsgTpt.T) == 0 {
			return t, errors.New("MsgTpt not init!")
		}
		msg0 := MsgTpt.T[0].MsgName
		index := 0
		for w := 0; w < len(t); w++ {
			if t[w].Value == msg0 {
				index = w
				log.Debug("Find the FeatureRaw start point is index=", index)
				break
			}
		}
		return t[index:], nil

	}
	return t, nil
}

//CaptureFeaturesSlice capture the feature raw from a file
//if full false, we just capture from the msg0
func CaptureFeaturesSlice(file string, full bool) ([]FeatureRaw, error) {

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

	t := make([]FeatureRaw, TptSize, TptCap)
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
			//log.Debug("Split each line  to =", r)
			if len(r) != 0 {
				for _, v := range r {
					if _, ok := MsgMap.M[v]; ok {
						var raw FeatureRaw
						raw.Value = v
						j := strings.Index(string(buffer), v)

						_, err := f.Seek(lpos+int64(j+len(v)), 0)
						if err != nil {
							return nil, err
						}
						rt, err := SeekTime(f, 500)
						if err != nil || len(rt) == 0 {
							log.Info("Can't find timestamp for->", v)
							t = append(t, raw)
							continue
						}
						w := strings.Replace(strings.Replace(rt, "[", "", -1), "]", "", -1)
						log.Debug("CaptureFeatures time:", w)
						raw.Time, raw.Nano, err = FindTimeNano(w)
						if err != nil {
							return nil, fmt.Errorf("FindTimeNano Error:%s", err)
						}
						t = append(t, raw)
					}
				}
			}
		}
	}

	//if full false, we need the msg0 as the start point
	if !full {
		if len(MsgTpt.T) == 0 {
			return t, errors.New("MsgTpt not init!")
		}
		msg0 := MsgTpt.T[0].MsgName
		index := 0
		for w := 0; w < len(t); w++ {
			if t[w].Value == msg0 {
				index = w
				log.Debug("Find the FeatureRaw start point is index=", index)
				break
			}
		}
		return t[index:], nil

	}
	return t, nil
}

func SeekTimeSlam(file *os.File, size int64) ([]string, error) {
	var logs []string
	rp := regexp.MustCompile(regx)
	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	sr := rp.FindAllIndex(buffer, -1)
	if sr != nil {
		for i := 0; i < len(sr); i++ {
			var log []byte
			if i == len(sr)-1 {
				log = buffer[sr[i][0]:]
			} else {
				log = buffer[sr[i][0]:(sr[i+1][0] - 1)]
			}
			logs = append(logs, string(log))
		}
		return logs, nil
	}

	return nil, errors.New("Empty TimeStamp Strings")
}

//SeekTime search the most close timestamp before current position
//TODO: make the split more precise, such as use the timestamp to split the files' lines
//if with the fixed size, it still will make the wrong timestamp
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
		log.Debug("npos=", npos, " size=", size)
		length, err := file.ReadAt(buffer, npos)
		if err != nil {
			return "", err
		}
		log.Debugf("Read %d buffer=>%s", length, buffer)
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

func PickFirstTime(t string) (string, error) {
	var result string
	var err error
	rp := regexp.MustCompile(regx)
	sr := rp.FindAllString(t, -1)
	if sr != nil {
		result = sr[0]
		log.Debug("Found closest timestamp", result)
	} else {
		err = errors.New("No TimeStamp in the string")
	}
	return result, err
}

func FindTimeNano(t string) (time.Time, int64, error) {
	s := strings.Split(t, ".")
	if len(s) != 2 {
		return time.Time{}, 0, errors.New("Wrong time format FindTimeNano")
	}
	tn, err := time.ParseInLocation("2006-01-02 15:04:05", s[0], time.Local)
	if err != nil {
		return time.Time{}, 0, err
	}

	na, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}

	return tn, na, nil
}

func PureDuplicate(c []FeatureRaw) []FeatureRaw {
	fnew := make([]FeatureRaw, TptSize, TptCap)

	mapit := make(map[FeatureRaw]bool)

	for _, v := range c {
		if _, ok := mapit[v]; !ok {
			mapit[v] = true
			fnew = append(fnew, v)
		}
	}
	return fnew
}

//TransformFeaturePure make the featureraw to featurepure
func TransformFeaturePure(c []FeatureRaw) ([]FeaturePure, error) {
	m0 := c[0]
	log.Debug("TransformFeaturePureMsg0=", m0.Value)

	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Time == c[j].Time {
			return c[i].Nano < c[j].Nano
		} else {
			return c[i].Time.Before(c[j].Time)
		}
	})

	//var cn FeatureRawChain = c
	//cn.Print()

	if m0.Value != c[0].Value {
		log.Debug("TransformFeaturePureMsg0=", c[0].Value)
		return nil, errors.New("Wrong Message0")
	}

	fp := make([]FeaturePure, TptSize, TptCap)

	for i := 0; i < len(c); i++ {
		var vn FeaturePure
		vn.Time = c[i].Time
		vn.Nano = c[i].Nano
		vn.Flag = c[i].Flag
		vn.Value = c[i].Value
		if i == 0 {
			vn.BoTime = 0
			vn.DeTime = 0
		} else {
			vn.BoTime = ComputeTimeDelta(c[0], c[i])
			vn.DeTime = ComputeTimeDelta(c[i-1], c[i])
		}

		if vn.BoTime < 0 || vn.DeTime < 0 {
			return nil, errors.New("TransformFeaturePure: lessthan0 error")
		}

		fp = append(fp, vn)
	}
	return fp, nil
}

func ComputeTimeDelta(before, after FeatureRaw) int64 {
	if before.Time.Before(after.Time) {
		d := after.Time.Sub(before.Time)
		log.Debug("before", before, after)
		return d.Nanoseconds() + after.Nano - before.Nano //TODO: how tomake the pricise detla?
	} else if before.Time.After(after.Time) {
		//not right if negative
		log.Debug("after", before, after)
		return -1
	} else {
		log.Debug("equal", before, after)
		return after.Nano - before.Nano
	}
	return -1
}
