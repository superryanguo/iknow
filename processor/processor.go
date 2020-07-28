package processor

import (
	"fmt"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

const (
	MaxMatch = 100
	Dec      = ".dec"
)

//BuildSvmTrainData build the output file with the input train data folder
//the file's name will show its svm class:+1(PositiveSample)/-1
func BuildSvmTrainData(input, output, tmpt string) error {
	var err error
	//tmptfile := "train/" + tmpt
	//log.Debug("BuildSvmTrainData TmptFilename=", tmptfile)

	feature.MsgTpt, err = feature.ExtractFeatureTemplate(tmpt)
	if err != nil {
		return err
	}
	feature.MsgTpt.Print()
	feature.MsgMap.Build(feature.MsgTpt)
	feature.MsgMap.Print()

	//1st step: check how many *.dec log in the folder
	decfiles, err := utils.FilterFileList(input, Dec)
	if err != nil {
		return err
	}

	//2nd step: capture the features and covert it to svm feature
	for k, v := range decfiles {
		log.Debug("BuildSvmTrainData_", k, ":", v)
		fr, err := feature.CaptureFeautres("./testdata/" + v)
		if err != nil {
			return err
		}
		feature.FeatureRawChain(fr).Print()
		fp, err := feature.TransformFeaturePure(feature.PureDuplicate(fr))
		if err != nil {
			return err
		}
		feature.FeaturePureChain(fp).Print()
	}

	//libSvm.No
	return nil

}

//TemplateMatch compare the resutl with the template, see if they match
//If the template message point =0, we don't support to match it, pls just
//Del the point 0 message in the template if it can't have a point
//if any circule has one match, then return true
//msg0 in the template is a key
func TemplateMatch(s feature.FeatureTestStatus, t feature.FeatureTemplate) (bool, error) {

	var index [MaxMatch]int //no more than 100
	var length int = 0      //Makesure
	var err error
	//first we need find how many begin message, which is the msg0
	//search the first index
	for k, v := range s.S {
		if v.MsgName == t.T[0].MsgName {
			index[length] = k
			length++
		}
	}

	if length == 0 {
		return false, fmt.Errorf("Can't find the msg0(%s) in the log", t.T[0].MsgName)
	}

	for i := 0; i < length; i++ {
		id := index[i]
		seq0 := s.S[id].Seq
		var find bool = true

		for j := 0; j < len(t.T); j++ {
			if t.T[j].Seq != (s.S[id].Seq-seq0+1) || t.T[j].MsgName != s.S[id].MsgName {
				log.Debugf("Mismatch: Seq=%d----%d,MsgName=%s-----%s\n", t.T[j].Seq, s.S[id].Seq-seq0+1, t.T[j].MsgName, s.S[id].MsgName)
				err = fmt.Errorf("id=%d, seq/message mismatch", j)
				find = false
				break
			}
			id++
		}

		if find {
			log.Debug("MatchFindStartId=", index[i])
			return true, nil
		}

	}

	return false, err

}
