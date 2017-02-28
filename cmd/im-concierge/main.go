package main

import (
	"fmt"
	"github.com/sapiens-sapide/IM-concierge/cmd/im-concierge/cli-cmds"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
