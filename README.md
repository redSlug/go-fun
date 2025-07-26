# go_fun
Command line tool to render the visual history of a node project from all commits over time

*Requirements*
- run in a Github repo with less than 100 commits
- `npm i` has already run and all required packages are included in the first commit
- `npm run dev` should bring up frontend on `http://localhost:5173/` with hot reloading enabled
- must have `ffmpeg` installed

## Develop
```bash
# install packages
go mod tidy

# run 
go run .
```

## Use
extract as a command line tool
```bash
# create the binary, -o gives it a unique name
go build -o go_fun

# create a symlink so it can be used anywhere
ln -s ~/Development/go_fun/go_fun /usr/local/bin/go_fun

# run from a github repo that allows `npm run dev`
go_fun
```

## Uses
- [bubbletea](https://github.com/charmbracelet/bubbletea) for command line interface
- [git-go](https://github.com/go-git/go-git) for interacting w/ git

## Next steps
- stitch together the screenshots and create a video from them
- handle case where packages are added in later commits
- consider shutting down and re-running the `npm run dev` command for each commit 
- obviate the need for `kill $(lsof -ti :5174)`
- fix the spacing of the output + hide red herring error messages
- get the count of commits and increment num for the SCRRENSHOT file name so they are not reversed
- create a video gif

## Troubleshooting
- `ERROR: could not unmarshal event: json: cannot unmarshal JSON string into Go network.IPAddressSpace within "/resourceIPAddressSpace": unknown IPAddressSpace value: Local` is a red herring
