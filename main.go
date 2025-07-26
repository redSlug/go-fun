package main

import (
	"context"
	"crypto/sha256"
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
	"strings"
	"time"
)

type model struct {
	pressed string
	error   string
}

const TestURL string = "http://localhost:5173/"
const ScreenshotNumDigits = 3
const FramesPerSecond = 2

func initialModel() model {
	return model{
		pressed: "",
		error:   "",
	}
}

func createUniqueFiles(prefix string, newPrefix string) {
	hasher := sha256.New()
	fileHashes := map[string]bool{}
	uniqueFilesPaths := []string{}
	fileNum := 0

	entries, _ := os.ReadDir(".")

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			content, _ := os.ReadFile(entry.Name())
			checksum := string(hasher.Sum(content))
			_, found := fileHashes[checksum]
			if !found {
				uniqueFilesPaths = append(uniqueFilesPaths, entry.Name())
				fileHashes[checksum] = true
				fileName := fmt.Sprintf("%s_%0*d.jpg", newPrefix, ScreenshotNumDigits, fileNum)
				fileNum++
				os.WriteFile(fileName, content, 0644)
			}
		}
	}

}

func createVideoFromUniqueFiles() {
	inputFilePattern := "UNIQUE_SCREENSHOT_%03d.jpg"
	outputFile := "UNIQUE_RENDER_HISTORY.mov"

	cmd := exec.Command(
		"ffmpeg",
		"-framerate", "2",
		"-pix_fmt", "rgb24",
		"-i", inputFilePattern,
		outputFile,
	)

	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	err := cmd.Run()

	if err != nil {
		log.Fatal("error running ffmpeg", err)
	}
}

func checkoutMain(worktree git.Worktree) {
	worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Main,
	})
}

func createScreenshotsFromCommits(m model, imageQuality int) tea.Cmd {
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

		imageFileExtension := "png"
		if imageQuality < 100 {
			imageFileExtension = "jpg"
		}

		var buf []byte
		captureError := chromedp.Run(ctx,
			chromedp.Navigate(TestURL),
			chromedp.Sleep(100*time.Millisecond),
			chromedp.EmulateViewport(1000, 500),
			chromedp.FullScreenshot(&buf, imageQuality),
		)

		if captureError != nil {
			log.Println("error capturing screenshot" + captureError.Error())
		}

		var fileName = fmt.Sprintf("SCREENSHOT_%d.%s", c.Author.When.Unix(), imageFileExtension)
		if err := os.WriteFile(fileName, buf, 0644); err != nil {
			checkoutMain(*worktree)
			log.Fatal("error writing file", err)
		}

		cancel()
		return nil
	})
	if iteratorError != nil {
		checkoutMain(*worktree)
		log.Fatal("chromedp error" + iteratorError.Error())
	}

	cancelParentServer()
	checkoutMain(*worktree)
	m.pressed = "PROCESSED COMMITS"
	return nil
}

func (m model) Init() tea.Cmd {
	createScreenshotsFromCommits(m, 80)
	createUniqueFiles("SCREENSHOT_", "UNIQUE_SCREENSHOT")
	createVideoFromUniqueFiles()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			createScreenshotsFromCommits(m, 80)
		case "u":
			createUniqueFiles("SCREENSHOT_", "UNIQUE_SCREENSHOT")
		case "v":
			createVideoFromUniqueFiles()
		case "q":
			return m, tea.Quit

		default:
			m.pressed = msg.String()
		}
	}
	return m, nil
}

func (m model) View() string {
	return "you pressed: " + m.pressed
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
