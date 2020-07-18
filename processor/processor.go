package main

import (
	"fmt"

	"github.com/superryanguo/iknow/config"
)

func main() {
	config.Init()
	fmt.Println("host:", config.GetProgramConfig().GetHost())
}
