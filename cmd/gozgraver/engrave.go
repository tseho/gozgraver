package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images
	_ "image/jpeg"

	// Package image/png is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand PNG formatted images
	_ "image/png"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	gozgraver "github.com/tseho/gozgraver/core"
)

func printUsageOfEngrave() {
	fmt.Print(`Usage:
    gozgraver <com> engrave <image> [<options>]

Arguments:
    com        COM address, eg: /dev/ttyUSB0 on Unix, COM45 on Windows
    image      path to the image file

Options:
        --burn int   Burn time in ms (default 18)
        --times int  Amount of passes (default 1)
        --power int  Laser power in percentage [1, 100] (default 60)
    -v, --verbose    Increase the verbosity
        --debug      Include all the traces in the logs
        --help       Display this help message
`)
}

func engrave() error {
	com := os.Args[1]

	// Define command and options
	cmd := flag.NewFlagSet("engrave", flag.ContinueOnError)
	burn := cmd.Int("burn", 18, "")
	times := cmd.Int("times", 1, "")
	power := cmd.Int("power", 60, "")
	v := cmd.BoolP("verbose", "v", false, "")
	debug := cmd.Bool("debug", false, "")

	// Silence the default error output of Parse()
	cmd.SetOutput(ioutil.Discard)

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		// when --help is passed as an option, display the usage
		// of the command instead of an error.
		if err.Error() == "pflag: help requested" {
			printUsageOfEngrave()
			return nil
		}

		fmt.Println(err.Error())
		return err
	}

	cmdArgs := cmd.Args()
	if len(cmdArgs) != 2 {
		fmt.Printf("invalid amount of arguments, expected 2, got %d", len(cmdArgs))
		return fmt.Errorf("invalid amount of arguments")
	}

	// If --verbose, set the log level to debug
	if *v {
		setLogLevel(logrus.DebugLevel)
	}

	// If --debug, set the log level to trace
	if *debug {
		setLogLevel(logrus.TraceLevel)
	}

	// Path to the image to be engraved
	path := cmdArgs[1]

	fmt.Printf("%s engrave %s %d %d %t\n", com, path, *burn, *times, *v)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	//Always be sure to have the file pointer at 0;0 before decoding
	file.Seek(0, 0)
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Open the graver connection
	graver, err := gozgraver.Connect(com)
	if err != nil {
		fmt.Println(err)
		return err
	}

	graver.SetBurnTime(*burn)
	graver.SetLaserPower(*power)
	err = graver.Engrave(img, *times)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
