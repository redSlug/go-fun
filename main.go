package main

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"log"
	"os"
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

func (m model) Init() tea.Cmd {
	dir, err := os.Getwd()
	if err != nil {
		m.error = "something went wrong" + dir
		log.Fatal(err)
	}

	repo, gitErr := git.PlainOpen(dir)
	if gitErr != nil {
		if errors.Is(gitErr, git.ErrRepositoryNotExists) {
			m.error = "Git repository does not exist" + dir

		}
		log.Fatal(gitErr)
	}

	ref, refError := repo.Head()
	if refError != nil {
		m.error = "git cannot get head ref"
	}

	commits, commitsError := repo.Log(&git.LogOptions{From: ref.Hash()})
	if commitsError != nil || commits == nil {
		m.error = "git cannot get commits"
	}

	worktree, worktreeError := repo.Worktree()
	if worktreeError != nil {
		m.error = "git cannot get worktree"
	}

	iteratorError := commits.ForEach(func(c *object.Commit) error {
		log.Println(c.Message)
		worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(c.Hash.String()),
		})

		return nil
	})
	if iteratorError != nil {
		m.error = iteratorError.Error()
		return nil
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
