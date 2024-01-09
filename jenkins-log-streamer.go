package main

import (
	"fmt"
	"github.com/jashort/jenkins-log-streamer/internal"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Usage: "Stream console log from a Jenkins job",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Value: "",
				Usage: "Jenkins job `URL`",
			},
			&cli.StringFlag{
				Name:    "user",
				Value:   "",
				Usage:   "Jenkins user",
				EnvVars: []string{"JENKINS_USER"},
			},
			&cli.StringFlag{
				Name:    "token",
				Value:   "",
				Usage:   "Jenkins API token",
				EnvVars: []string{"JENKINS_TOKEN"},
			},
		},
		Action: func(cCtx *cli.Context) error {

			server := jenkins.ServerInfo{
				URL:   cCtx.String("url"),
				User:  cCtx.String("user"),
				Token: cCtx.String("token"),
			}

			jobStatus := jenkins.FetchJobStatus(server)
			fmt.Printf("Name: %s\n", jobStatus.FullDisplayName)
			fmt.Printf("Start time: %d\n", jobStatus.Timestamp)
			fmt.Printf("Result: %s\n", jobStatus.Result)
			fmt.Printf("Building: %t\n", jobStatus.Building)
			fmt.Printf("In Progress: %t\n\n", jobStatus.InProgress)

			process := func(data string) bool {
				if len(data) > 0 {
					fmt.Print(data)
				}
				return false
			}

			jenkins.FetchLog(server, process)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
