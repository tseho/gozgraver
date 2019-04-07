package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	gozgraver "github.com/tseho/gozgraver/core"
)

func usage() {
	fmt.Print(`Usage:
    gozgraver <com> <command>

Arguments:
    com        COM address, eg: /dev/ttyUSB0 on Unix, COM45 on Windows
    command    one of the commands described below

Commands:
    engrave    engrave an image
    reset      reset the graver to default settings

Use "gozgraver <com> <command> --help" for more information about a specific command.
`)
}

func setLogLevel(level logrus.Level) {
	gozgraver.GetLogger().SetLevel(level)
}

func main() {
	// default logs level to info
	setLogLevel(logrus.InfoLevel)

	if len(os.Args) < 3 {
		usage()
		return
	}

	cmd := os.Args[2]

	switch cmd {
	case "engrave":
		engrave()
	case "reset":
		reset()
	default:
		fmt.Printf("Unknown command %s", cmd)
		usage()
	}
}
