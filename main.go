package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ftsctl/models"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

func main() {

	ftsUrlFlag := flag.String("url", "foo", "The url of FTSnext")
	flag.Parse()

	u, err := url.ParseRequestURI(*ftsUrlFlag)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("url: ", u)

	statusUrl, err := startProcess(u, "example")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(statusUrl)

	m := model{
		progress:  progress.New(progress.WithDefaultGradient()),
		statusUrl: statusUrl,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

func fetchProcessStatus(url string) (models.ProcessStatus, error) {
	resp, err := http.Get(url)
	if err != nil {
		return models.ProcessStatus{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var status models.ProcessStatus
	err = json.Unmarshal([]byte(body), &status)
	if err != nil {
		log.Fatal(err)
	}
	return status, nil
}

func startProcess(baseUrl *url.URL, project string) (string, error) {
	req, err := http.NewRequest(
		"POST",
		baseUrl.JoinPath("/api/v2/process/"+project+"/start").String(),
		nil,
	)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	statusUrl := resp.Header.Get("Content-Location")
	if statusUrl == "" {
		log.Println("No Content-Location")
	}

	return statusUrl, nil
}

type tickMsg time.Time

type model struct {
	progress  progress.Model
	statusUrl string
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		status, err := fetchProcessStatus(m.statusUrl)
		if err != nil {
			log.Fatal(err)
		}

		if status.Phase == "COMPLETED" && m.progress.Percent() == 1.0 {
			return m, tea.Quit
		}

		if status.DeidentifiedBundles > 0 {
			cmd := m.progress.SetPercent(float64(status.SentBundles+status.SkippedBundles) / float64(status.DeidentifiedBundles))
			return m, tea.Batch(tickCmd(), cmd)
		} else {
			return m, tea.Batch(tickCmd(), nil)
		}

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press any key to quit")
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
