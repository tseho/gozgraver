package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	gozgraver "github.com/tseho/gozgraver/core"
)

func printUsageOfReset() {
	fmt.Print(`Usage:
    gozgraver <com> reset

Arguments:
    com        COM address, eg: /dev/ttyUSB0 on Unix, COM45 on Windows

Options:
    -v, --verbose    Increase the verbosity
        --debug      Include all the traces in the logs
        --help       Display this help message
`)
}

func reset() error {
	com := os.Args[1]

	// Define command and options
	cmd := flag.NewFlagSet("reset", flag.ContinueOnError)
	v := cmd.BoolP("verbose", "v", false, "")
	debug := cmd.Bool("debug", false, "")

	// Silence the default error output of Parse()
	cmd.SetOutput(ioutil.Discard)

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		// when --help is passed as an option, display the usage
		// of the command instead of an error.
		if err.Error() == "pflag: help requested" {
			printUsageOfReset()
			return nil
		}

		fmt.Println(err.Error())
		return err
	}

	// If --verbose, set the log level to debug
	if *v {
		setLogLevel(logrus.DebugLevel)
	}

	// If --debug, set the log level to trace
	if *debug {
		setLogLevel(logrus.TraceLevel)
	}

	// Open the graver connection
	graver, err := gozgraver.Connect(com)
	if err != nil {
		fmt.Println(err)
		return err
	}

	graver.Reset()

	return nil
}
