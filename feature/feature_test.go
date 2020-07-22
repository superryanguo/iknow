package feature

import (
	"fmt"
	"os"
	"testing"
)

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
