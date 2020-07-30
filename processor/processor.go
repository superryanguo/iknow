package processor

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

const (
	MaxMatch      = 100
	Dec           = ".dec"
	PosClass      = "+1" //Else name
	NegClass      = "-1" //If name has NegativeClass string
	NagtiveSymbol = "NegativeClass"
	MatrixD       = 3
)

//BuildSvmTrainData build the output file with the input train data folder
//the file's name will show its svm class:+1(PositiveSample)/-1
func BuildSvmTrainData(input, output, tmpt string) error {
	var err error
	var label string
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

	if utils.CheckFileExist(output) {
		log.Debug("FileExist, remove", output, "first")
		err := os.Remove(output)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	//2nd step: capture the features and covert it to svm feature
	for k, v := range decfiles {
		log.Debug("BuildSvmTrainData_", k, ":", input+"/"+v)
		fr, err := feature.CaptureFeautres(input+"/"+v, false)
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
		//log.Debug("fpn:", fpn)

		if find := strings.Contains(v, NagtiveSymbol); find {
			label = NegClass
		} else {
			label = PosClass
		}
		log.Debug("Sample Lable:", label)
		r, err := MapFeatPurFullToDeSvmFloat(fpn, feature.MsgTpt)
		if err != nil {
			return err
		}
		log.Debug("SVM Feature:", label+" "+r)
		_, err = f.WriteString(label + " " + r + "\n")
		if err != nil {
			return err
		}

	}

	return nil

}

//MapFeatPurFullToBoSvm map the FeaturePure to SVM lib formate string
//we use the full feature to descibe the msgexist/seq/time
//template will the matrix mother template, each msg 3 features
func MapFeatPurFullToBoSvm(fr feature.FeaturePureChain, t feature.FeatureTemplate) (string, error) {
	var result string
	var index int = 1

	for i := 0; i < len(t.T); i++ {
		msg := t.T[i].MsgName
		var r string
		for k, v := range fr {
			if v.Value == msg { //message exist
				r += strconv.Itoa(MatrixD*i+index) + ":" + "1" + " "                 //1 means message exist
				r += strconv.Itoa(MatrixD*i+index+1) + ":" + strconv.Itoa(k+1) + " " //seq number
				if v.BoTime == 0 {
					r += strconv.Itoa(MatrixD*i+index+2) + ":" + "1" + " " //TODO: Ajust Detime 0 to 1, will it be good?
				} else {
					r += strconv.Itoa(MatrixD*i+index+2) + ":" + strconv.FormatInt(v.BoTime, 10) + " " //Detime
				}
			}
		}
		result += r
	}
	return result, nil
}

//TODO: how to hanlde circles in the feature pure, such as msg0-msg3-msg0-msg5,
//we need to split into serveral svm feature, but not to combine it into one
func CircleSplitFeature() {

}

func MapFeatPurFullToDeSvmFloat(fr feature.FeaturePureChain, t feature.FeatureTemplate) (string, error) {
	var result string
	var index int = 1

	for i := 0; i < len(t.T); i++ {
		msg := t.T[i].MsgName
		var r string
		for k, v := range fr {
			if v.Value == msg { //message exist
				r += strconv.Itoa(MatrixD*i+index) + ":" + "1" + " "                             //1 means message exist
				r += strconv.Itoa(MatrixD*i+index+1) + ":" + strconv.Itoa(k+1) + " "             //seq number
				r += strconv.Itoa(MatrixD*i+index+2) + ":" + fmt.Sprintf("%.6f", v.NorVal) + " " //Detime
			}
		}
		result += r
	}
	return result, nil
}

//MapFeatPurFullToDeSvm map the FeaturePure to SVM lib formate string
//we use the full feature to descibe the msgexist/seq/time
//template will the matrix mother template, each msg 3 features
func MapFeatPurFullToDeSvm(fr feature.FeaturePureChain, t feature.FeatureTemplate) (string, error) {
	var result string
	var index int = 1

	for i := 0; i < len(t.T); i++ {
		msg := t.T[i].MsgName
		var r string
		for k, v := range fr {
			if v.Value == msg { //message exist
				r += strconv.Itoa(MatrixD*i+index) + ":" + "1" + " "                 //1 means message exist
				r += strconv.Itoa(MatrixD*i+index+1) + ":" + strconv.Itoa(k+1) + " " //seq number
				if v.DeTime == 0 {
					r += strconv.Itoa(MatrixD*i+index+2) + ":" + "1" + " " //TODO: Ajust Detime 0 to 1, will it be good?
				} else {
					r += strconv.Itoa(MatrixD*i+index+2) + ":" + strconv.FormatInt(v.DeTime, 10) + " " //Detime
				}
			}
		}
		result += r
	}
	return result, nil
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
