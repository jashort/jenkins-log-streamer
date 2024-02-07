package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	jenkins "github.com/jashort/jenkins-log-streamer/internal"
	"github.com/jashort/jenkins-log-streamer/internal/jlsviewport"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"time"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.NormalBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.NormalBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type model struct {
	// Program state
	server   jenkins.ServerInfo
	ready    bool
	viewport jlsviewport.Model
	content  string
	debug    bool
	// Jenkins job state
	jobStartTime    int64
	jobName         string
	jobStatus       string
	err             error
	job             jenkins.JobStatus
	secondsLeft     int
	currentBuildNum int
	logPosition     int64
	moreData        bool
	logChunks       []logChunk
}

func (m model) headerView() string {
	statusLine := ""
	if m.jobStatus != "" {
		statusLine = "[" + m.jobStatus + "]"
	}
	startTime := time.UnixMilli(m.jobStartTime).Format(time.RFC822)
	fmtLine := "%s %s (Started %s)"
	//Log Position: %d   More data: %t    Refresh in: %d`

	title := titleStyle.Render(fmt.Sprintf(fmtLine, m.jobName, statusLine, startTime))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("Refresh in %d        %3.f%%", m.secondsLeft, m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type logChunk struct {
	lineCount int
	lines     []string
}

type jobStatusMsg struct {
	name       string
	startTime  int64
	buildNum   int
	inProgress bool
	result     string
}

type jobLogMsg struct {
	start       int64
	body        string
	moreData    bool
	newPosition int64
	buildNum    int
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), updateStatus(m.server), tea.EnterAltScreen)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	if m.debug {
		log.Println(fmt.Sprintf("(%T): %s\n", message, message))
	}
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = jlsviewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case jobStatusMsg:
		m.jobStartTime = msg.startTime
		m.jobName = msg.name
		if msg.result != "" {
			m.jobStatus = msg.result
		} else {
			if msg.inProgress {
				m.jobStatus = "In Progress"
			}
		}

		// If the latest build number has changed, clear the log
		if m.currentBuildNum != msg.buildNum {
			m.logChunks = nil
			m.logPosition = 0
			m.currentBuildNum = msg.buildNum
			m.moreData = true
			m.content = ""
		}

		if m.moreData {
			return m, updateLog(m.server, m.logPosition, m.currentBuildNum)
		} else {
			return m, nil
		}

	case jobLogMsg:
		if msg.buildNum == m.currentBuildNum {
			if len(msg.body) != 0 {
				shouldScroll := m.viewport.AtBottom()
				lines := strings.Split(msg.body, "\n")
				chunk := logChunk{
					lineCount: len(lines),
					lines:     lines,
				}
				m.logChunks = append(m.logChunks, chunk)
				m.content += msg.body
				m.viewport.SetContent(m.content)
				if shouldScroll {
					m.viewport.GotoBottom()
				}
			}

			m.logPosition = msg.newPosition
			m.moreData = msg.moreData
			// moreData may mean "the log is finished, but you're not at the last chunk" or it
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

	m.viewport, cmd = m.viewport.Update(message)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func updateStatus(server jenkins.ServerInfo) tea.Cmd {
	return func() tea.Msg {
		response, err := jenkins.FetchJobStatus(server)
		if err != nil {
			return errMsg{err}
		}
		x := jobStatusMsg{
			name:       response.FullDisplayName,
			startTime:  response.Timestamp,
			buildNum:   response.Number,
			inProgress: response.InProgress,
			result:     response.Result,
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
			&cli.StringFlag{
				Name:    "log",
				Value:   "",
				Usage:   "Log debugging information to filename",
				EnvVars: []string{"JLS_LOG"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			debugMode := false
			if cCtx.String("log") != "" {
				if _, err := tea.LogToFile(cCtx.String("log"), ""); err != nil {
					log.Fatal(err)
				}
				debugMode = true
			}
			server := jenkins.ServerInfo{
				JobBaseUrl: cCtx.String("url"),
				User:       cCtx.String("user"),
				Token:      cCtx.String("token"),
			}
			if server.JobBaseUrl == "" {
				log.Fatal("Error: jenkins URL not specified. Use --url option")
			}
			p := tea.NewProgram(
				model{secondsLeft: 5, server: server, debug: debugMode},
				tea.WithAltScreen(),
			)
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
