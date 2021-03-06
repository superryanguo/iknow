package processor

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

const (
	MaxMatch              = 100
	Dec                   = ".dec"
	PosValue      float64 = 1
	NegValue      float64 = -1
	PosClass              = "+1" //Else name
	NegClass              = "-1" //If name has NegativeClass string
	NagtiveSymbol         = "NegativeClass"
	MatrixD               = 3
	NegF                  = "NegF"
	PosF                  = "PosF"
)

//BuildSvmTrainData build the output file with the trainpath train data folder
//the file's name will show its svm class:+1(PositiveSample)/-1
//if trainfiles is not empty, we will use it as the training input
//or just use the files under the trainpath
func BuildSvmTrainData(trainpath, output, tmpt string, trainfiles []string) error {
	var err error
	var label string
	var decfiles []string
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

	if utils.CheckFileExist(output) {
		log.Debug("FileExist, remove ", output, " first")
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

	if len(trainfiles) == 0 { //TODO: to be improved
		decfiles, err = TrainFileCollect(trainpath)
		if err != nil {
			return err
		}
	} else {
		decfiles = trainfiles
	}

	log.Debug("TrainFileList:", decfiles)
	if len(decfiles) == 0 {
		return errors.New("0 train files found")
	}

	//2nd step: capture the features and covert it to svm feature
	for k, v := range decfiles {
		log.Debug("BuildSvmTrainData_", k, ":", trainpath+"/"+v)
		var fr []feature.FeatureRaw
		//if len(trainfiles) == 0 { //TODO: to be improved
		//fr, err = feature.CaptureFeatures(trainpath+"/"+v, false)
		//} else {
		_, fr, err = feature.CaptureFeatures(v, false)
		//}
		if err != nil {
			return err
		}
		feature.FeatureRawChain(fr).Print()
		//TODO: That's how we handle the empty feautre list,
		//how about we just write a labe in SVM file, will it be good?
		if len(fr) == 0 {
			label = NegClass
			_, err = f.WriteString(label + " " + "\n")
			if err != nil {
				return err
			}
			log.Info("the file:", v, " has empty feature list")
			continue
		}
		fp, err := feature.TransformFeaturePure(feature.PureDuplicate(fr))
		if err != nil {
			return err
		}
		feature.FeaturePureChain(fp).Print()
		var fpn feature.FeaturePureChain = fp
		fpn.SvmDeTimeNormalize(1, 2)
		//log.Debug("fpn:", fpn)

		if find := strings.Contains(v, NagtiveSymbol) || strings.Contains(v, NegF); find {
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
	var result []string
	var index int = 1
	var match []int
	length := len(fr)

	//message0 is the first index
	for k, v := range fr {
		if v.Value == t.T[0].MsgName {
			match = append(match, k)
		}
	}
	log.Debug("Get the indexmatch: ", match)
	if len(match) == 0 {
		return MapFeatPurFullToDeSvmFloatV1(fr, t)
	}

	for _, w := range match {
		var r string
		j := w
		mask := make([]int, length, length)
		for i := 0; i < len(t.T); i++ {
			if j >= len(fr) {
				break
			}

			if fr[j].Value == t.T[i].MsgName {
				r += strconv.Itoa(MatrixD*i+index) + ":" + "1" + " "                                 //1 means message exist
				r += strconv.Itoa(MatrixD*i+index+1) + ":" + strconv.Itoa(j-w+1) + " "               //seq number
				r += strconv.Itoa(MatrixD*i+index+2) + ":" + fmt.Sprintf("%.6f", fr[j].NorVal) + " " //Detime
				mask[j] = 1
				j++
			} else {
				for h := j + 1; h < len(fr); h++ {
					if fr[h].Value == t.T[i].MsgName && mask[h] != 1 {
						//j = h TODO: not good to reset the j index?! should add mask about the message has been used to avoid duplicate name
						mask[h] = 1                                                                          //used
						r += strconv.Itoa(MatrixD*i+index) + ":" + "1" + " "                                 //1 means message exist
						r += strconv.Itoa(MatrixD*i+index+1) + ":" + strconv.Itoa(h-w+1) + " "               //seq number
						r += strconv.Itoa(MatrixD*i+index+2) + ":" + fmt.Sprintf("%.6f", fr[h].NorVal) + " " //Detime
					}
				}
			}
		}
		log.Debugf("w=%d, mask=%v\n", w, mask)
		result = append(result, r)
	}

	if len(result) == 0 {
		return "", errors.New("No SVM FeaVector")
	}
	log.Debug("ResultSum: ", result)
	n := 0
	for x, y := range result {
		if len(y) > len(result[n]) {
			n = x
		}
	}

	return result[n], nil
}
func MapFeatPurFullToDeSvmFloatV1(fr feature.FeaturePureChain, t feature.FeatureTemplate) (string, error) {
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

//TemplateMatch compare the result with the template, see if they match
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
			if id >= len(s.S) {
				log.Debug("test file message short")
				err = fmt.Errorf("id=%d, seq/message mismatch", j+1)
				find = false
				break
			}
			if t.T[j].Seq != (s.S[id].Seq-seq0+1) || t.T[j].MsgName != s.S[id].MsgName {
				log.Debugf("Mismatch: Seq=%d----%d,MsgName=%s-----%s\n", t.T[j].Seq, s.S[id].Seq-seq0+1, t.T[j].MsgName, s.S[id].MsgName)
				err = fmt.Errorf("id=%d, seq/message mismatch", j+1)
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

func TrainFileCollect(datap string) ([]string, error) {
	testPos := datap + "/" + PosF
	testNeg := datap + "/" + NegF
	var decfiles []string
	dirs := []string{testNeg, testPos}
	for _, dir := range dirs {
		d, err := utils.FilterFileList(dir, Dec)
		if err != nil {
			return nil, err
		}
		decfiles = append(decfiles, d...)
	}
	return decfiles, nil

}
