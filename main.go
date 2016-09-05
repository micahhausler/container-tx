package main

import (
	"fmt"
	"os"

	"github.com/micahhausler/container-tx/transform"
	flag "github.com/ogier/pflag"
)

const Version = "0.0.1"

var version = flag.Bool("version", false, "print version and exit")

var inputType = flag.StringP("input", "i", "compose", "The format of the input")
var outputType = flag.StringP("output", "o", "ecs", "The format of the output")
var file = flag.StringP("file", "f", "docker-compose.yaml", "The input file")

var inputMap = map[string]transform.InputFormat{
	"compose": transform.ComposeFormat{},
	"ecs":     transform.EcsFormat{},
}

var outputMap = map[string]transform.OutputFormat{
	"compose": transform.ComposeFormat{},
	"ecs":     transform.EcsFormat{},
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Printf("container-tx %s\n", Version)
		os.Exit(0)
	}

	input, ok := inputMap[*inputType]
	if !ok {
		inputKeys := []string{}
		for it := range inputMap {
			inputKeys = append(inputKeys, it)
		}
		fmt.Printf("Input type %s invalid: must be one of %s\n", *inputType, inputKeys)
		os.Exit(1)
	}
	f, err := os.Open(*file)
	basePod, err := input.IngestContainers(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outputFormat, ok := outputMap[*outputType]
	if !ok {
		outputKeys := []string{}
		for it := range inputMap {
			outputKeys = append(outputKeys, it)
		}
		fmt.Printf("Output type %s invalid: must be one of %s\n", *outputType, outputKeys)
		os.Exit(1)
	}
	resp, err := outputFormat.EmitContainers(basePod)

	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
}
