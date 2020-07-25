package processor

import (
	"fmt"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/config"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

var logfile = "./testdata/sGnb.log"

const (
	MaxMatch = 100
)

func Process() {
	config.Init()
	//log.Info("Host:", config.GetProgramConfig().GetHost())
	log.Info("TemplateFilePath:", config.GetProgramConfig().GetTemplatePath())
	file := config.GetProgramConfig().GetTemplatePath()
	if !utils.CheckFileExist(file) {
		log.Info("TemplateFileNotExist!")
		return
	}

	var err error
	feature.MsgTpt, err = feature.ExtractFeatureTemplate(file)
	if err != nil {
		log.Fatal(err)
	}
	feature.MsgTpt.Print()
	feature.MsgMap.Build(feature.MsgTpt)
	feature.MsgMap.Print()
	l, err := feature.CaptureFeautres(logfile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("+++++++++++++++")
	fmt.Println(l)
	fmt.Println("+++++++++++++++")

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
