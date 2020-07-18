package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-micro/v2/config/source/file"
	log "github.com/micro/go-micro/v2/logger"
)

var (
	err error
)

var (
	defaultRootPath         = "app"
	defaultConfigFilePrefix = "application-"
	programConfig           defaultProgramConfig
	profiles                defaultProfiles
	m                       sync.RWMutex
	inited                  bool
)

type Profiles interface {
	GetInclude() string
}

type defaultProfiles struct {
	Include string `json:"include"`
}

func (p defaultProfiles) GetInclude() string {
	return p.Include
}

func Init() {
	m.Lock()
	defer m.Unlock()

	if inited {
		log.Infof("[Init] Already init the configuration")
		return
	}

	appPath, _ := filepath.Abs(filepath.Dir(filepath.Join("./", string(filepath.Separator))))

	pt := filepath.Join(appPath, "conf")
	os.Chdir(appPath)

	if err = config.Load(file.NewSource(file.WithPath(pt + "/application.yml"))); err != nil {
		panic(err)
	}

	if err = config.Get(defaultRootPath, "profiles").Scan(&profiles); err != nil {
		panic(err)
	}

	log.Infof("[Init] The configuration files：%s, %+v\n", pt+"/application.yml", profiles)

	if len(profiles.GetInclude()) > 0 {
		include := strings.Split(profiles.GetInclude(), ",")

		sources := make([]source.Source, len(include))
		for i := 0; i < len(include); i++ {
			filePath := pt + string(filepath.Separator) + defaultConfigFilePrefix + strings.TrimSpace(include[i]) + ".yml"

			log.Infof("[Init] loading：%s\n", filePath)

			sources[i] = file.NewSource(file.WithPath(filePath))
		}

		if err = config.Load(sources...); err != nil {
			panic(err)
		}
	}

	config.Get(defaultRootPath, "program").Scan(&programConfig)

	inited = true
}

func GetProgramConfig() (ret ProgramConfig) {
	return programConfig
}
