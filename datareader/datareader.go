package datareader

import (
	"encoding/gob"
	"flag"
	"os"

	log "github.com/micro/go-micro/v2/logger"
)

//TODO:
//1.make it ./datareader <filepath>

type MesgData struct {
	Name    string
	Summary string
}

type AccRecd struct {
	RecData map[string][]string
}

func (a *AccRecd) ReadFile(f string) error {
	_, e := os.Stat(f)
	if e != nil {
		if os.IsNotExist(e) {
			log.Info("File not exist")
			return e
		} else { //shouldn't go here
			log.Warn(e)
		}
	} else {
		file, err := os.Open(f)
		defer file.Close()
		if err != nil {
			return err
		}
		err = gob.NewDecoder(file).Decode(&a.RecData)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func (a *AccRecd) ShowData() {
	for k, v := range a.RecData {
		for i, s := range v {
			log.Infof("|%s|%d|%s|\n", k, i, s)
		}

	}
}

var DataLib AccRecd

func init() {
	//log.SetOutput(os.Stdout)
	//log.SetLevel(log.InfoLevel)
	//log.SetLevel(log.DebugLevel)
	DataLib.RecData = make(map[string][]string)
}

func main() {
	file := flag.String("file", "../datastore/data.gb", "Data file path")
	flag.Parse()
	log.Debug("KickOff the DataReader...", *file)
	err := DataLib.ReadFile(*file)
	if err != nil {
		log.Warn(err)
		return
	}
	log.Debug("DataStore readfile...")
	DataLib.ShowData()
}
