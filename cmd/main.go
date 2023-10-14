package main

import (
	"Yi/pkg/runner"
)

/**
  @author: yhy
  @since: 2022/10/13
  @desc: //TODO
**/

func main() {
	runner.ParseArguments()
	runner.Init()
	runner.Run()

	if runner.Option.Target == "" && runner.Option.Targets == "" {
		go runner.Init()
	}
}
