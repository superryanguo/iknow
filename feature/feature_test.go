package feature

import (
	"fmt"
	"os"
	"testing"
)

func TestSeekTimeSlam(t *testing.T) {
	fn := "../processor/testdata/sGnb.log"

	f, err := os.Open(fn)
	if err != nil {
		t.Error("SeekTimeSlamErr:", err)
		return
	}

	defer f.Close()

	r, err := SeekTimeSlam(f, 100)
	if err != nil {
		t.Error("SeekTimeSlamErr:", err)
		return
	}

	for k, v := range r {
		fmt.Println(k, "->", v)
		fmt.Println("")
	}
}

func TestSeekTime(t *testing.T) {
	fn := "../processor/testdata/sGnb.log"

	f, err := os.Open(fn)
	if err != nil {
		t.Error("SeekTimeErr:", err)
		return
	}

	defer f.Close()

	pos, err := f.Seek(800, 0)
	if err != nil {
		t.Error("SeekTimeErr:", err)
		return
	}
	fmt.Println("pos=", pos)
	r, err := SeekTime(f, 100)
	if err != nil {
		t.Error("SeekTimeErr:", err)
		return
	}
	fmt.Println("SeekTime find the timestamp", r)

}

func TestFindTimeNano(t *testing.T) {
	ti := "2019-12-04 06:00:21.486673325"
	var nano int64 = 486673325
	time := "2019-12-04 06:00:21"
	a, b, e := FindTimeNano(ti)
	if e != nil {
		t.Error("FindTimeNano:", e)
		return
	}
	if nano != b {
		t.Error("FindTimeNano:Nano Mismatch")
		return
	}

	if a.Format("2006-01-02 15:04:05") != time {
		t.Error("FindTimeNano:time Mismatch")
		return
	}
	fmt.Println("Target:", ti)
	fmt.Println("Get the right value", a, b)

}

//TODO: Make this testcaputre func better to test a full process of feature's package
func TestCaputre(t *testing.T) {
	fn := "../processor/testdata/sGnb.log"
	tptfile := "../processor/testdata/ho.tmpt"

	//TODO: why not the same as feature.MsgTpt?!! bug? the feature package not share the global value in test package?!
	MsgTpt, err := ExtractFeatureTemplate(tptfile)
	if err != nil {
		t.Error("TestCaputre:", err)
		return
	}
	MsgTpt.Print()
	MsgMap.Build(MsgTpt)
	MsgMap.Print()
	_, l, err := CaptureFeatures(fn, false)
	if err != nil {
		t.Error("TestCaputre:", err)
		return
	}
	var c FeatureRawChain = l
	c.Print()
	newl := PureDuplicate(l)
	c = newl
	c.Print()
	p, err := TransformFeaturePure(c)
	if err != nil {
		t.Error("TestCaputre:", err)
		return
	}
	var cp FeaturePureChain = p
	cp.Print()
}
