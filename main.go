package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"log"
	"os"
	"os/exec"
	"rsc.io/quote"
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

func (m model) Init() tea.Cmd {
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
			}
			cmd := exec.Command("npm", "run", "dev")

			output, execError := cmd.CombinedOutput()
			if execError != nil {
				log.Fatal("npm run error" + execError.Error() + "\n" + string(output))
			}
		}
		return nil
	})
	if iteratorError != nil {
		log.Fatal("iterator error" + iteratorError.Error())
	}
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
