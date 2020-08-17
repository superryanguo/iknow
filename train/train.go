package main

import (
	"flag"

	log "github.com/micro/go-micro/v2/logger"

	libSvm "github.com/ewalker544/libsvm-go"
	"github.com/superryanguo/iknow/processor"
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
		TestBenchmark(*hotype)
		log.Info("test flag")
	} else if *usage == "hybird" {
		log.Info("hybird flag")
	} else {
		log.Info("Unsupported flag")
	}

}
