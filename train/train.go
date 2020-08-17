package main

import (
	"errors"
	"flag"

	log "github.com/micro/go-micro/v2/logger"

	libSvm "github.com/ewalker544/libsvm-go"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/learning"
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
	NegF          = "Neg"
	PosF          = "Pos"
)

func TrainModel(t string) {
	svmpara := libSvm.NewParameter()
	svmpara.KernelType = libSvm.POLY
	model := libSvm.NewModel(svmpara)

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

	file := traindata + "/" + Trainfile
	log.Debug("trained_data_file=", file)
	err := processor.BuildSvmTrainData(traindata, file, traintmpt)
	if err != nil {
		log.Fatal("BuildSvmTrainDataErr:", err)
	}
	problem, err := libSvm.NewProblem(file, svmpara)
	if err != nil {
		log.Info("ProblemCreateErr:", err)
		return
	}

	err = model.Train(problem)
	if err != nil {
		log.Fatal("ModelTrainErr:", err)
	}

	err = model.Dump(trainmodel)
	if err != nil {
		log.Fatal("ModelTrainErr:", err)
	}

}

func TestBenchmark(t string) error {
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
		errors.New("Unsupport data type")
	}
	testPos := testdata + "/" + PosF
	testNeg := testdata + "/" + NegF
	var numPos, numvPos, numNeg, numvNeg, numTotal int
	var perNeg, perPos, perVerify float64
	var err error

	numPos, err = utils.CountFileNum(testPos)
	if err != nil {
		return err
	}
	numNeg, err = utils.CountFileNum(testNeg)
	if err != nil {
		return err
	}
	numTotal = numPos + numNeg
	if numTotal == 0 {
		return errors.New("No test file found")
	}
	log.Info("numPos=", numPos, "numNeg=", numNeg)

	feature.MsgTpt, err = feature.ExtractFeatureTemplate(traintmpt)
	if err != nil {
		return err
	}
	feature.MsgTpt.Print()
	feature.MsgMap.Build(feature.MsgTpt)
	feature.MsgMap.Print()

	dirs := []string{testNeg, testPos}
	for _, dir := range dirs {

		decfiles, err := utils.FilterFileList(dir, processor.Dec)
		if err != nil {
			return err
		}

		for _, v := range decfiles {
			fr, err := feature.CaptureFeatures(dir+"/"+v, false)
			if err != nil {
				return err
			}
			feature.FeatureRawChain(fr).Print()
			fp, err := feature.TransformFeaturePure(feature.PureDuplicate(fr))
			if err != nil {
				return err
			}
			feature.FeaturePureChain(fp).Print()
			var fpn feature.FeaturePureChain = fp
			fpn.SvmDeTimeNormalize(1, 2)
			ml, err := learning.SvmLearn(trainmodel, fpn, feature.MsgTpt)
			if err != nil {
				return err
			}
			//mlr := fmt.Sprintf("%f", ml)
			log.Debug("mlr=", ml)
			if ml == processor.PosValue {
				if dir == testPos {
					numvPos++
				}
			} else if ml == processor.NegValue {
				if dir == testNeg {
					numvNeg++
				}
			} else {
				return errors.New("Unknown learning result label")
			}
		}
	}

	perNeg = float64(numvNeg) / float64(numNeg)
	perPos = float64(numvPos) / float64(numPos)
	perVerify = float64(numvPos+numvNeg) / float64(numTotal)
	log.Infof("Result: numvNeg=%d,numNeg=%d,numPos=%d,numvPos=%d\n", numvNeg, numNeg, numPos, numvPos)
	log.Infof("Result: perNeg=%f,perPos=%f,perVerify=%f\n", perNeg, perPos, perVerify)

	return nil
}

func main() {
	hotype := flag.String("hotype", "hotgt", "HandoverType: hotgt, hosrc")
	//hybird will rotate the all data in the train and test folder to do train and test.
	usage := flag.String("usage", "train", "Usage: train, test, hybird")
	flag.Parse()

	log.Debug("flag input=", *hotype, "|", *usage)
	if *usage == "train" {
		TrainModel(*hotype)
	} else if *usage == "test" {
		err := TestBenchmark(*hotype)
		if err != nil {
			log.Fatal(err)
		}
	} else if *usage == "hybird" {
		log.Info("hybird flag")
	} else {
		log.Info("Unsupported flag")
	}

}
