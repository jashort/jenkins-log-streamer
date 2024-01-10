package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"time"
)

const url = "https://httpbin.org/delay/3"

type model struct {
	url          string
	user         string
	token        string
	jobStartTime int64
	jobName      string
	jobStatus    string
	err          error
	secondsLeft  int
}

type jobStatusMsg struct {
	name      string
	startTime int64
}

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), updateStatus(), tea.EnterAltScreen)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case jobStatusMsg:
		m.jobStartTime = msg.startTime
		m.jobName = msg.name

	case tickMsg:
		m.secondsLeft--
		if m.secondsLeft <= 0 {
			return m, tea.Quit
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("%s %s\n%d\n\nSeconds left: %d", m.jobName, m.jobStatus, m.jobStartTime, m.secondsLeft)
}

func updateStatus() tea.Cmd {
	return func() tea.Msg {
		c := &http.Client{
			Timeout: 10 * time.Second,
		}
		res, err := c.Get(url)
		if err != nil {
			return errMsg{err}
		}
		defer res.Body.Close() // nolint:errcheck

		return tea.Msg(jobStatusMsg{name: "Test", startTime: 3})

	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

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

			//server := jenkins.ServerInfo{
			//	URL:   cCtx.String("url"),
			//	User:  cCtx.String("user"),
			//	Token: cCtx.String("token"),
			//}

			p := tea.NewProgram(model{secondsLeft: 5}, tea.WithAltScreen(), tea.WithFPS(30))
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
			//	jobStatus := jenkins.FetchJobStatus(server)
			//	fmt.Printf("Name: %s\n", jobStatus.FullDisplayName)
			//	fmt.Printf("Start time: %d\n", jobStatus.Timestamp)
			//	fmt.Printf("Result: %s\n", jobStatus.Result)
			//	fmt.Printf("Building: %t\n", jobStatus.Building)
			//	fmt.Printf("In Progress: %t\n\n", jobStatus.InProgress)
			//
			//	process := func(data string) bool {
			//		if len(data) > 0 {
			//			fmt.Print(data)
			//		}
			//		return false
			//	}
			//
			//	jenkins.FetchLog(server, process)
			//
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
