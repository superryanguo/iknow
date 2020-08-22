package main

import (
	"errors"
	"flag"
	"math/rand"
	"time"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/processor"
	"github.com/superryanguo/iknow/utils"
)

const (
	SrcHoTestPath = "test_data/HoSrc"
	TgtHoTestPath = "test_data/HoTgt"
	SrcHoDataPath = "train_data/HoSrc"
	TgtHoDataPath = "train_data/HoTgt"
	SrcHoTmptPath = "train_tempt/HoSrc.tmpt"
	TgtHoTmptPath = "train_tempt/HoTgt.tmpt"
	SrcHoModel    = "train_model/HoSrc.Model"
	TgtHoModel    = "train_model/HoTgt.Model"
	Trainfile     = "svm.train"
)

func TrainModel(t string) {

	var traindata, traintmpt, trainmodel string
	if t == "hotgt" {
		traindata = TgtHoDataPath
		traintmpt = TgtHoTmptPath
		trainmodel = TgtHoModel
	} else if t == "hosrc" {
		traindata = SrcHoDataPath
		traintmpt = SrcHoTmptPath
		trainmodel = SrcHoModel
	} else {
		log.Fatal("Unsupport data type")
	}
	err := TrainSvModel(traindata, traintmpt, trainmodel, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func BenchmarkModel(t string) error {
	var testdata, traintmpt, trainmodel string
	if t == "hotgt" {
		testdata = TgtHoTestPath
		traintmpt = TgtHoTmptPath
		trainmodel = TgtHoModel
	} else if t == "hosrc" {
		testdata = SrcHoTestPath
		traintmpt = SrcHoTmptPath
		trainmodel = SrcHoModel
	} else {
		return errors.New("Unsupport data type")
	}
	r, err := BenchmarkSvModel(testdata, traintmpt, trainmodel, nil)
	if err != nil {
		return err
	}
	r.Print()
	s := CalPercent(r)
	s.Print()
	return nil
}

func HybirdBenchmark(t string, tper float64) error {
	var traindata, testdata, traintmpt, trainmodel string
	var err error
	if t == "hotgt" {
		testdata = TgtHoTestPath
		traindata = TgtHoDataPath
		traintmpt = TgtHoTmptPath
		trainmodel = TgtHoModel
	} else if t == "hosrc" {
		traindata = SrcHoDataPath
		testdata = SrcHoTestPath
		traintmpt = SrcHoTmptPath
		trainmodel = SrcHoModel
	} else {
		return errors.New("Unsupport data type")
	}

	if tper >= 0.5 || tper <= 0 {
		return errors.New("TestPercent must >0 and <0.5")
	}
	testPos := testdata + "/" + processor.PosF
	testNeg := testdata + "/" + processor.NegF
	//1st collect the files from the test and train folder
	var filesum []string
	for _, v := range []string{testPos, testNeg, traindata} {
		fl, err := utils.FilterFileList(v, processor.Dec)
		if err != nil {
			return err
		}
		filesum = append(filesum, fl...)
	}
	log.Debug("CombileFileList:", filesum)

	rand.Seed(time.Now().UnixNano())
	utils.Shuffle(filesum)
	log.Debug("ShuffleCombileFileList:", filesum)

	//2nd rotate train with the percentage files and test with other files
	total := len(filesum)
	testnum := int(float64(total) * tper)
	//trainnum := total - testnum
	var result []HybirdResult

	for i := 0; i < total; i++ {
		var tslice, aslice []string
		if i < total-testnum {
			log.Debug("Index:", i, "-ts:", i+testnum)
			tslice = append(tslice, filesum[i:i+testnum]...)
			aslice = append(append(aslice, filesum[0:i]...), filesum[i+testnum:]...)
		} else {
			log.Debug("Index:", i, "-ts:", i+testnum-total)
			tslice = append(append(tslice, filesum[i:]...), filesum[0:i+testnum-total]...)
			aslice = append(aslice, filesum[i+testnum-total:i]...)
		}
		log.Debugf("tslice:%v\n", tslice)
		log.Debugf("aslice:%v\n", aslice)
		err = TrainSvModel(traindata, traintmpt, trainmodel, aslice)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("i=", i, "train done->", trainmodel)
		r, err := BenchmarkSvModel(testdata, traintmpt, trainmodel, tslice)
		if err != nil {
			return err
		}
		r.Print()
		s := CalPercent(r)
		s.Print()
		result = append(result, HybirdResult{s, aslice, tslice})
		log.Info("i=", i, "test done")
	}

	for k, v := range result {
		log.Infof("the %dth Result=%s\n", k, v.Percent.ToString())
	}
	return nil
}

func main() {
	hotype := flag.String("hotype", "hotgt", "HandoverType: hotgt, hosrc")
	//hybird will rotate the all data in the train and test folder to do train and test.
	usage := flag.String("usage", "train", "Usage: train, test, hybird, train-auto")
	tper := flag.Float64("tper", 0.2, "Usage: 0<tper<0.5")
	flag.Parse()

	log.Debug("flag input=", *hotype, "|", *usage)
	if *usage == "train" {
		TrainModel(*hotype)
	} else if *usage == "test" {
		err := BenchmarkModel(*hotype)
		if err != nil {
			log.Fatal(err)
		}
	} else if *usage == "hybird" {
		HybirdBenchmark(*hotype, *tper)
		log.Info("hybird flag", *tper)
	} else if *usage == "train-auto" {
		log.Info("auto train flag")
	} else {
		log.Info("Unsupported flag")
	}

}
