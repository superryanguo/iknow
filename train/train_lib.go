package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/micro/go-micro/v2/logger"

	libSvm "github.com/ewalker544/libsvm-go"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/learning"
	"github.com/superryanguo/iknow/processor"
	"github.com/superryanguo/iknow/utils"
)

type ModelResult struct {
	NumNeg  int
	NumvNeg int
	NumPos  int
	NumvPos int
}

type Percent struct {
	PerNeg    float64
	PerPos    float64
	PerVerify float64
}

type HybirdResult struct {
	Percent
	Aslice []string
	Tslice []string
}

func (r ModelResult) Print() {
	log.Infof("\nModelResult:\nnumvNeg=%d\nnumNeg=%d\nnumPos=%d\nnumvPos=%d\n", r.NumvNeg, r.NumNeg, r.NumPos, r.NumvPos)
}

func (r ModelResult) ToString() string {
	return fmt.Sprintf("\nModelResult:\nnumvNeg=%d\nnumNeg=%d\nnumPos=%d\nnumvPos=%d\n", r.NumvNeg, r.NumNeg, r.NumPos, r.NumvPos)
}
func (p Percent) ToString() string {
	return fmt.Sprintf("\nPercent Result:\nperNeg=%0.2f%%\nperPos=%0.2f%%\nperVerify=%0.2f%%\n", p.PerNeg*100, p.PerPos*100, p.PerVerify*100)
}
func (p Percent) Print() {
	log.Infof("\nPercent Result:\nperNeg=%0.2f%%\nperPos=%0.2f%%\nperVerify=%0.2f%%\n", p.PerNeg*100, p.PerPos*100, p.PerVerify*100)
}

func CalPercent(m ModelResult) (p Percent) {
	total := m.NumPos + m.NumNeg
	if total == 0 {
		log.Info("CalPercent total=0!")
		return
	}
	if m.NumNeg != 0 {
		p.PerNeg = float64(m.NumvNeg) / float64(m.NumNeg)
	}

	if m.NumPos != 0 {
		p.PerPos = float64(m.NumvPos) / float64(m.NumPos)
	}

	p.PerVerify = float64(m.NumvPos+m.NumvNeg) / float64(total)

	return
}

func BenchmarkSvModel(datap, tmpt, model string, trainfiles []string) (ModelResult, error) {
	var numPos, numvPos, numNeg, numvNeg, numTotal int
	var mr ModelResult
	var err error

	if len(datap) == 0 {
		return mr, errors.New("Empty DataPath found in BenchmarkSvModel")
	}
	testPos := datap + "/" + processor.PosF
	testNeg := datap + "/" + processor.NegF

	feature.MsgTpt, err = feature.ExtractFeatureTemplate(tmpt)
	if err != nil {
		return mr, err
	}
	feature.MsgTpt.Print()
	feature.MsgMap.Build(feature.MsgTpt)
	feature.MsgMap.Print()

	if len(trainfiles) == 0 {
		numPos, err = utils.CountFileNum(testPos)
		if err != nil {
			return mr, err
		}
		numNeg, err = utils.CountFileNum(testNeg)
		if err != nil {
			return mr, err
		}
		numTotal = numPos + numNeg
		if numTotal == 0 {
			return mr, errors.New("No test file found")

		}
	} else {
		numTotal = len(trainfiles)
		for _, f := range trainfiles {
			if strings.Contains(f, processor.NagtiveSymbol) || strings.Contains(f, processor.NegF) {
				numNeg++
			}
		}
		numPos = numTotal - numNeg
	}
	mr.NumPos = numPos
	mr.NumNeg = numNeg
	log.Info("numPos=", numPos, "numNeg=", numNeg)

	var decfiles []string

	if len(trainfiles) == 0 {
		dirs := []string{testNeg, testPos}
		for _, dir := range dirs {
			d, err := utils.FilterFileList(dir, processor.Dec)
			if err != nil {
				return mr, err
			}
			decfiles = append(decfiles, d...)
		}
	} else {
		decfiles = trainfiles
	}

	for _, v := range decfiles {
		fr, err := feature.CaptureFeatures(v, false)
		if err != nil {
			return mr, err
		}
		feature.FeatureRawChain(fr).Print()
		//TODO: That's how we handle the empty feautre list,
		//how about we just write a labe in SVM file, will it be good?
		if len(fr) == 0 {
			ml, err := learning.SvmLearn(model, nil, feature.MsgTpt)
			if err != nil {
				return mr, err
			}
			log.Info("benchmark the file:", v, " has empty feature list, svm result:", ml)
			continue
		}
		fp, err := feature.TransformFeaturePure(feature.PureDuplicate(fr))
		if err != nil {
			return mr, err
		}
		feature.FeaturePureChain(fp).Print()
		var fpn feature.FeaturePureChain = fp
		fpn.SvmDeTimeNormalize(1, 2)
		ml, err := learning.SvmLearn(model, fpn, feature.MsgTpt)
		if err != nil {
			return mr, err
		}
		//mlr := fmt.Sprintf("%f", ml)
		log.Debug("mlr=", ml)
		if ml == processor.PosValue {
			if !strings.Contains(v, processor.NagtiveSymbol) && !strings.Contains(v, processor.NegF) {
				numvPos++
			}
		} else if ml == processor.NegValue {
			if strings.Contains(v, processor.NagtiveSymbol) || strings.Contains(v, processor.NegF) {
				numvNeg++
			}
		} else {
			return mr, errors.New("Unknown learning result label")
		}
	}

	mr.NumvNeg = numvNeg
	mr.NumvPos = numvPos
	return mr, nil

}

func TrainSvModel(datap, tmpt, output string, trainfiles []string) error {
	var err error
	svmpara := libSvm.NewParameter()
	svmpara.KernelType = libSvm.POLY
	model := libSvm.NewModel(svmpara)

	if len(datap) == 0 {
		return errors.New("wrong datapath for the model training")
	}
	file := datap + "/" + Trainfile
	log.Debug("trained_data_file=", file)

	if len(trainfiles) != 0 { //TODO: to be improved
		err = processor.BuildSvmTrainData(datap, file, tmpt, trainfiles)
	} else {
		err = processor.BuildSvmTrainData(datap, file, tmpt, nil)
	}
	if err != nil {
		return errors.New(fmt.Sprintf("BuildSvmTrainDataErr:%s", err))
	}
	problem, err := libSvm.NewProblem(file, svmpara)
	if err != nil {
		return errors.New(fmt.Sprintf("ProblemCreateErr:", err))
	}

	err = model.Train(problem)
	if err != nil {
		return errors.New(fmt.Sprintf("ModelTrainErr:", err))
	}

	if utils.CheckFileExist(output) {
		log.Debug("FileExist, remove", output, "first")
		err = os.Remove(output)
		if err != nil {
			return err
		}
	}

	err = model.Dump(output)
	if err != nil {
		return errors.New(fmt.Sprintf("ModelTrainErr:", err))
	}
	return nil
}
