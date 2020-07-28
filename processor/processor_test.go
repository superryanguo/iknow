package processor

import (
	"testing"
)

func TestBuildSvmTrainData(t *testing.T) {
	dir := "./testdata/"
	tmpt := "./testdata/ho.tmpt"
	out := "./testdata/svm.train"
	err := BuildSvmTrainData(dir, out, tmpt)

	if err != nil {
		t.Error("BuildSvmTrainData:", err)
		return
	}

}
