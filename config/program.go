package config

type ProgramConfig interface {
	GetEnabled() bool
	GetPort() int
	GetHost() string
	GetAlg() string
	GetTemplatePath() string
}

type defaultProgramConfig struct {
	Enabled      bool   `json:"enabled"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Alg          string `json:"alg"`
	Templatepath string `json:"templatepath"`
}

func (c defaultProgramConfig) GetTemplatePath() string {
	return c.Templatepath
}

func (c defaultProgramConfig) GetPort() int {
	return c.Port
}

func (c defaultProgramConfig) GetEnabled() bool {
	return c.Enabled
}

func (c defaultProgramConfig) GetHost() string {
	return c.Host
}

func (c defaultProgramConfig) GetAlg() string {
	return c.Alg
}
