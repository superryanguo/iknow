package learning

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	libSvm "github.com/ewalker544/libsvm-go"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/superryanguo/iknow/feature"
	"github.com/superryanguo/iknow/processor"
)

func SvmLearn(md string, fr feature.FeaturePureChain, t feature.FeatureTemplate) (float64, error) {

	model := libSvm.NewModelFromFile(md)

	//r, err := processor.MapFeatPurFullToDeSvm(fr, t)
	r, err := processor.MapFeatPurFullToDeSvmFloat(fr, t)
	if err != nil {
		return 0, err
	}
	log.Debug("SvmLearn r:=", r)
	x, err := PopulateResultToX(r)
	if err != nil {
		return 0, err
	}

	log.Debug("SvmLearn TestVector x:=", x)
	predictLabel := model.Predict(x) // Predicts a float64 label given the test vector
	log.Debug("Result=", predictLabel)
	return predictLabel, nil
}

func PopulateResultToX(s string) (map[int]float64, error) {
	x := make(map[int]float64)
	var err error
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return nil, errors.New("wrong capture feature format")
	}

	for _, v := range tokens {
		if len(v) > 0 {
			node := strings.Split(v, ":")
			if len(node) > 1 {
				var index int
				var value float64
				if index, err = strconv.Atoi(node[0]); err != nil {
					return nil, fmt.Errorf("Fail to parse index from token %v\n", v)
				}
				if value, err = strconv.ParseFloat(node[1], 64); err != nil {
					return nil, fmt.Errorf("Fail to parse value from token %v\n", v)
				}
				x[index] = value
			}
		}
	}
	return x, nil
}
