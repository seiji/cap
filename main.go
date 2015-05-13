package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var (
	version string = "HEAD"
)

func main() {
	newApp().Run(os.Args)
}

func newApp() (app *cli.App) {
	app = cli.NewApp()
	app.Name = "cap"
	app.Usage = "deploy script"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "stage, s",
			Value: "prod",
			Usage: "stage for deploy",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "init",
			Aliases:     []string{"i"},
			Usage:     "create config files",
			Action: capInit,
		},
		{
			Name:      "deploy",
			Aliases:     []string{"d"},
			Usage:     "add a task to the list",
			Action: capDeploy,
		},
		{
			Name:      "setup",
			Aliases:     []string{"s"},
			Usage:     "complete a task on the list",
			Action: capSetup,
		},
	}
	return
}
