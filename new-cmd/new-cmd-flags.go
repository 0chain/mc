package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/codegangsta/cli"
)

func _assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseInput(c *cli.Context) {
	var commandName string
	switch len(c.Args()) {
	case 1:
		commandName = c.Args()[0]
	default:
		log.Fatal("command name must not be blank\n")
	}

	var inputOptions []string
	if c.String("options") != "" {
		inputOptions = strings.Split(c.String("options"), ",")
	}

	if inputOptions[0] == "" {
		log.Fatal("options cannot be empty with a command name")
	}

	var commandUsage string
	if c.String("usage") != "" {
		commandUsage = c.String("usage")
	}

	var mainObject = template.Must(template.New("main").Parse(commandTemplate))
	var optionsObject = template.Must(template.New("options").Parse(optionsTemplate))
	var readmeObject = template.Must(template.New("readme").Parse(readmeTemplate))

	err := os.Mkdir(commandName, 0755)
	_assert(err)

	command := initCommand(commandName, commandUsage, inputOptions)

	optionsGo := source{
		Name:     commandName + "-options.go",
		TempLate: *optionsObject,
	}

	readmeMd := source{
		Name:     commandName + ".md",
		TempLate: *readmeObject,
	}

	mainGo := source{
		Name:     commandName + ".go",
		TempLate: *mainObject,
	}

	err = readmeMd.get(commandName, command)
	_assert(err)

	mainGo.get(commandName, command)
	_assert(err)

	optionsGo.get(commandName, command)

	err = GoFormat(commandName)
	_assert(err)
}
