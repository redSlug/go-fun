package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chromedp/chromedp"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"rsc.io/quote"
	"time"
)

type model struct {
	pressed string
	error   string
}

func initialModel() model {
	return model{
		pressed: "",
		error:   "",
	}
}

var TestCommitMessage string = "add link to try out the app\n"
var TestURL string = "http://localhost:5173/japanese-learning-helper/"
var TestFile = "TESTFile.png"

func writeFile(content []byte) {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal("Error getting current user: " + err.Error())
	}
	homeDir := currentUser.HomeDir

	relativePath := "temp/" + TestFile

	fullPath := filepath.Join(homeDir, relativePath)

	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal("Error creating directory", err.Error())
		}
	}

	err = ioutil.WriteFile(fullPath, content, 0644)
	if err != nil {
		log.Fatal("Error writing to file" + err.Error())
	}
}

func processCommits() tea.Cmd {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	repo, gitErr := git.PlainOpen(dir)
	if gitErr != nil {
		log.Fatal("Git repository does not exist in directory")
	}

	ref, refError := repo.Head()
	if refError != nil {
		log.Fatal("git cannot get head ref")
	}

	commits, commitsError := repo.Log(&git.LogOptions{From: ref.Hash()})
	if commitsError != nil || commits == nil {
		log.Fatal("git cannot get commits")
	}

	worktree, worktreeError := repo.Worktree()
	if worktreeError != nil {
		log.Fatal("git cannot get worktree")
	}

	iteratorError := commits.ForEach(func(c *object.Commit) error {
		if c.Message == TestCommitMessage {
			checkoutErr := worktree.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(c.Hash.String()),
			})
			if checkoutErr != nil {
				log.Fatal("checkout error" + checkoutErr.Error())
				return nil
			}
			cmd := exec.Command("npm", "run", "dev")

			output, execError := cmd.CombinedOutput()
			if execError != nil {
				log.Fatal("npm run error" + execError.Error() + "\n" + string(output))
				return nil
			}

			ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
			defer cancel()

			ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			var buf []byte

			captureError := chromedp.Run(ctx,
				chromedp.Navigate(TestURL),
				chromedp.CaptureScreenshot(&buf),
			)

			if captureError != nil {
				log.Fatal("error capturing screenshot" + captureError.Error())
			}

			writeFile(buf)
		}
		return nil
	})
	if iteratorError != nil {
		log.Fatal("iterator error" + iteratorError.Error())
	}
	return nil
}

func (m model) Init() tea.Cmd {
	m.pressed = "PRESS A KEY"
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "r":
			processCommits()
			return m, nil

		case "q":
			return m, tea.Quit

		default:
			m.pressed = msg.String()
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "you pressed: " + m.pressed
	return s
}

func main() {
	fmt.Println(quote.Go())
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("haark there was an error %v", err)
		os.Exit(1)
	}
	initialModel()
}
