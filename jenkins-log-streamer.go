package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	jenkins "github.com/jashort/jenkins-log-streamer/internal"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"time"
)

type model struct {
	jobStartTime    int64
	jobName         string
	jobStatus       string
	err             error
	job             jenkins.JobStatus
	server          jenkins.ServerInfo
	secondsLeft     int
	currentBuildNum int
	logPosition     int64
	moreData        bool
	logChunks       []logChunk
	programLog      []string
}

type logChunk struct {
	lineCount int
	lines     []string
}

type jobStatusMsg struct {
	name      string
	startTime int64
	buildNum  int
}

type jobLogMsg struct {
	start       int64
	body        string
	moreData    bool
	newPosition int64
	buildNum    int
}

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), updateStatus(m.server), tea.EnterAltScreen)
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
		// If the latest build number has changed, clear the log
		if m.currentBuildNum != msg.buildNum {
			m.logChunks = nil
			m.logPosition = 0
			m.currentBuildNum = msg.buildNum
			m.moreData = true
		}

		if m.moreData {
			return m, updateLog(m.server, m.logPosition, m.currentBuildNum)
		} else {
			return m, nil
		}

	case jobLogMsg:
		m.programLog = append(m.programLog, fmt.Sprintf("%#v", message))
		if msg.buildNum == m.currentBuildNum {
			if len(msg.body) != 0 {
				lines := strings.Split(msg.body, "\n")
				chunk := logChunk{
					lineCount: len(lines),
					lines:     lines,
				}
				m.logChunks = append(m.logChunks, chunk)
			}

			m.logPosition = msg.newPosition
			m.moreData = msg.moreData
			// moreData may mean "the log is finished but you're not at the last chunk" or it
			// may mean "the job is still running but there's no new data in the log". In the second
			// case, we don't want to immediately try to get more data, wait for updating the job
			// status to trigger it
			if msg.moreData && len(msg.body) > 0 {
				return m, updateLog(m.server, msg.newPosition, msg.buildNum)
			}
		}
		return m, nil

	case tickMsg:
		m.secondsLeft--
		if m.secondsLeft <= 0 {
			m.secondsLeft = 5
			return m, tea.Batch(updateStatus(m.server), tick())
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	//return strings.Join(m.programLog, "\n")
	logChunk := ""
	if len(m.logChunks) > 0 {
		curChunk := m.logChunks[len(m.logChunks)-1]
		start := 0
		end := 0
		if len(curChunk.lines) > 5 {
			start = len(curChunk.lines) - 5
			end = len(curChunk.lines)
		} else {
			end = len(curChunk.lines)
		}
		logChunk = strings.Join(curChunk.lines[start:end], "\n")
	}
	outFmt := `
	%s %s
	StartTime: %d
	Refresh in: %d	Log Position: %d More Data: %t
	Log:
	%s
	`
	return fmt.Sprintf(outFmt, m.jobName, m.jobStatus, m.jobStartTime, m.secondsLeft, m.logPosition, m.moreData, logChunk)
}

func updateStatus(server jenkins.ServerInfo) tea.Cmd {
	return func() tea.Msg {
		response, err := jenkins.FetchJobStatus(server)
		if err != nil {
			return errMsg{err}
		}
		x := jobStatusMsg{
			name:      response.FullDisplayName,
			startTime: response.Timestamp,
			buildNum:  response.Number,
		}
		return tea.Msg(x)
	}
}

func updateLog(server jenkins.ServerInfo, start int64, jobNumber int) tea.Cmd {
	return func() tea.Msg {
		data := jenkins.FetchLog(server, start)
		x := jobLogMsg{
			body:        data.Body,
			start:       start,
			newPosition: data.NewPosition,
			moreData:    data.MoreData,
			buildNum:    jobNumber,
		}
		return tea.Msg(x)
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
				Usage: "Jenkins job `Url`",
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
				JobBaseUrl: cCtx.String("url"),
				User:       cCtx.String("user"),
				Token:      cCtx.String("token"),
			}

			p := tea.NewProgram(model{secondsLeft: 5, server: server}, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
