package main

import (
	"github.com/knight42/kubectl-blame/cmd"
)

func main() {
	c := cmd.NewCmdBlame()
	_ = c.Execute()
}
