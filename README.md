# go_fun
command line tool to render the visual history of your frontend from all git commits over time

## Develop
install packages
```bash
go mod tidy
```

run
```bash
go run .
```

extract as a command line tool
```bash
# create the binary, -o gives it a unique name
go build -o go_fun

# create a symlink so it can be used anywhere
ln -s ~/Development/go_fun/go_fun /usr/local/bin/go_fun

```

## Uses
- [bubbletea](https://github.com/charmbracelet/bubbletea) for command line interface
- [git-go](https://github.com/go-git/go-git) for interacting w/ git

## Next steps
- checkout each commit, build and run, take a screenshot in headless broser, stitch together the screenshots
