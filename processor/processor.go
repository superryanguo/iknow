package main

import (
	"fmt"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/config"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/utils"
)

var logfile = "./testdata/sGnb.log"

func main() {
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
	l, err := feature.CaptureFeautresPlus(logfile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("+++++++++++++++")
	fmt.Println(l)
	fmt.Println("+++++++++++++++")

}
