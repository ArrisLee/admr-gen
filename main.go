package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ArrisLee/admr-gen/cmd"
)

func main() {
	var yamlFile, operation, output string
	flag.StringVar(&yamlFile, "file", "", "Path to the input YAML file")
	flag.StringVar(&operation, "operation", "", "Operation type (create, update, delete)")
	flag.StringVar(&output, "output", "", "Output type (yaml, json)")
	flag.Parse()

	params := &cmd.Params{
		YamlFile:  yamlFile,
		Operation: operation,
		Output:    output,
	}

	if err := params.Validate(); err != nil {
		log.Fatal(err)
	}

	outputStr, err := cmd.Run(params)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(outputStr)
}
