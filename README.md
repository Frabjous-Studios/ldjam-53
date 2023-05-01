# bankwave

A basic skeleton for a 2D game using the [Ebitengine](https://ebitengine.org/) library.

Extended from [the original](sinisterstuf/bankwave) with:
1) A credits screen
2) A main menu
3) YarnSpinner integration
4) Probably some other things too, but most importantly _no actual game logic_.

-----✂️-----

> ⚠️ After cloning this repository:

> Write your OWN name name in the LICENSE file and find-and-replace `bankwave` with your own game's name, for example, by running this command (tested on Linux and Mac):

```bash
grep -Rl bankwave | xargs sed -i '' -e "s/bankwave/${PWD##*/}/g"
```

> it assumes that the game name is the name of the current folder because that is what `go build` will call it.

> Then delete this section from the README, and start editing `main.go` to make your own game!

-----✂️-----

## For game testers

<!-- TODO: add a link to the latest downloads page -->

Game controls:
- F: toggle full-screen
- Q: quit the game
- Space: move up

## For programmers

Make sure you have [Go 1.20 or later](https://go.dev/) to contribute to the game

To build the game yourself, run: `go build .` it will produce an bankwave file and on Windows bankwave.exe.

To run the tests, run: `go test ./...` but there are no tests yet.

The project has a very simple, flat structure, the first place to start looking is the main.go file.
