package main

import (
	"bufio"
	"fmt"
	"os"

	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/config"
)

func main() {
	config.Init()
	//log.Info("Host:", config.GetProgramConfig().GetHost())
	log.Info("Templatepath:", config.GetProgramConfig().GetTemplatePath())
	file := config.GetProgramConfig().GetTemplatePath()

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		fmt.Println(s.Text())
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}
}
