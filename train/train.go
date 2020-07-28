package main

import (
	"flag"

	log "github.com/micro/go-micro/v2/logger"

	libSvm "github.com/ewalker544/libsvm-go"
	"github.com/superryanguo/iknow/processor"
)

const (
	SrcHoDataPath = "train_data/HoSrc"
	TgtHoDataPath = "train_data/HoTgt"
	SrcHoTmptPath = "train_tempt/HoSrc.tmpt"
	TgtHoTmptPath = "train_tempt/HoTgt.tmpt"
	SrcHoModel    = "train_model/HoSrc.Model"
	TgtHoModel    = "train_model/HoTgt.Model"
	Trainfile     = "svm.train"
)

func main() {
	hotype := flag.String("hotype", "hotgt", "HandoverType: hotgt, hosrc")
	flag.Parse()

	log.Debug("flag input=", *hotype)
	svmpara := libSvm.NewParameter()
	svmpara.KernelType = libSvm.POLY
	model := libSvm.NewModel(svmpara)

	var traindata, traintmpt, trainmodel string
	if *hotype == "hotgt" {
		traindata = TgtHoDataPath
		traintmpt = TgtHoTmptPath
		trainmodel = TgtHoModel
	} else if *hotype == "hosrc" {
		traindata = SrcHoDataPath
		traintmpt = SrcHoTmptPath
		trainmodel = SrcHoModel
	} else {
		log.Fatal("Unsupport data type")
	}

	file := traindata + "/" + Trainfile
	log.Debug("traindatafile=", file)
	err := processor.BuildSvmTrainData(traindata, file, traintmpt)
	if err != nil {
		log.Fatal("BuildSvmTrainDataErr:", err)
	}
	problem, err := libSvm.NewProblem(file, svmpara)
	if err != nil {
		log.Println("ProblemCreateErr:", err)
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
