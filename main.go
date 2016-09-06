package main

import (
	"fmt"
	"io"
	"os"

	"github.com/micahhausler/container-tx/compose"
	"github.com/micahhausler/container-tx/ecs"
	"github.com/micahhausler/container-tx/script"
	"github.com/micahhausler/container-tx/transform"
	flag "github.com/ogier/pflag"
)

const versionNum = "0.0.1"

var version = flag.Bool("version", false, "print version and exit")

var inputType = flag.StringP("input", "i", "compose", "The format of the input.")
var outputType = flag.StringP("output", "o", "ecs", "The format of the output.")

var inputMap = map[string]transform.InputFormat{
	"compose": compose.DockerCompose{},
	"ecs":     ecs.Task{},
}

var outputMap = map[string]transform.OutputFormat{
	"compose": compose.DockerCompose{},
	"ecs":     ecs.Task{},
	"cli":     script.Script{},
}

func main() {
	inputKeys := []string{}
	for it := range inputMap {
		inputKeys = append(inputKeys, it)
	}
	outputKeys := []string{}
	for it := range outputMap {
		outputKeys = append(outputKeys, it)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] <file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    Valid input types:  %s\n", inputKeys)
		fmt.Fprintf(os.Stderr, "    Valid output types: %s\n\n", outputKeys)
		fmt.Fprint(os.Stderr, "    If no file is specified, defaults to STDIN\n\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	if *version {
		fmt.Printf("container-tx %s\n", versionNum)
		os.Exit(0)
	}

	var f io.ReadCloser
	if fileName := flag.Arg(0); len(fileName) > 0 {
		var err error
		f, err = os.Open(fileName)
		if err != nil {
			fmt.Printf("Error opening file: %s \n", err)
			os.Exit(1)
		}
	} else {
		f = os.Stdin
	}

	input, ok := inputMap[*inputType]
	if !ok {
		fmt.Printf("Input type %s invalid: must be one of %s\n", *inputType, inputKeys)
		os.Exit(1)
	}

	outputFormat, ok := outputMap[*outputType]
	if !ok {
		fmt.Printf("Output type %s invalid: must be one of %s\n", *outputType, outputKeys)
		os.Exit(1)
	}

	basePod, err := input.IngestContainers(f)
	if err != nil {
		fmt.Printf("Error ingesting file: %s \n", err)
		os.Exit(1)
	}
	resp, err := outputFormat.EmitContainers(basePod)

	if err != nil {
		fmt.Printf("Error converting file: %s \n", err)
		os.Exit(1)
	}
	fmt.Println(string(resp))
}
