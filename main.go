package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chromedp/chromedp"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"log"
	"os"
	"os/exec"
	"rsc.io/quote"
	"time"
)

type model struct {
	pressed string
	error   string
}

var TestURL string = "http://localhost:5173/"

func initialModel() model {
	return model{
		pressed: "",
		error:   "",
	}
}

func checkoutMain(worktree git.Worktree) {
	worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Main,
	})
}

func processCommits(m model) tea.Cmd {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	repo, gitErr := git.PlainOpen(dir)
	if gitErr != nil {
		log.Fatal("git repository does not exist in directory")
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

	parentServerCtx, cancelParentServer := context.WithCancel(context.Background())
	defer cancelParentServer()

	cmd := exec.CommandContext(parentServerCtx, "npm", "run", "dev")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Println("could not start npm")
		return nil
	}
	time.Sleep(5 * time.Second)

	iteratorError := commits.ForEach(func(c *object.Commit) error {
		worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(c.Hash.String()),
		})

		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("ignore-certificate-errors", "1"),
			chromedp.Flag("allow-insecure-localhost", "1"),
			chromedp.Flag("disable-web-security", "1"),
		)

		ctx, cancel := chromedp.NewExecAllocator(parentServerCtx, opts...)
		defer cancel()

		ctx, cancel = chromedp.NewContext(ctx)
		defer cancel()

		var buf []byte
		captureError := chromedp.Run(ctx,
			chromedp.Navigate(TestURL),
			chromedp.Sleep(1*time.Second),
			chromedp.FullScreenshot(&buf, 80),
		)

		if captureError != nil {
			checkoutMain(*worktree)
			log.Fatal("error capturing screenshot" + captureError.Error())
		}

		if err := os.WriteFile("SCREENSHOT_"+c.Hash.String()+".png", buf, 0644); err != nil {
			log.Fatal("error writing file", err)
		}

		cancel()
		return nil
	})
	if iteratorError != nil {
		log.Fatal("iterator error" + iteratorError.Error())
	}

	cancelParentServer()
	checkoutMain(*worktree)
	m.pressed = "PROCESSED COMMITS"
	return nil
}

func (m model) Init() tea.Cmd {
	processCommits(m)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

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
