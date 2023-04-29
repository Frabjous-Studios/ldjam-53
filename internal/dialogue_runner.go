package internal

import (
	"embed"
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/DrJosh9000/yarn/bytecode"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/razor-1/localizer-cldr/resources/language"
	"sync"
)

const yarnFile = "gamedata/yarn/bin/game"

// yarnBin contains all yarn output from this compilation process.
//
//go:embed gamedata/yarn/bin
var yarnBin embed.FS

type RunnerState uint8

const (
	RunnerStopped RunnerState = iota // RunnerStopped
	RunnerRunning                    // RunnerRunning is set for a runner that's running.
	RunnerWaiting                    // RunnerWaiting indicates the runner is waiting for the player to select an dialogueOption.
)

// DialogueRunner runs YarnSpinner and any commands from the script. It buffers lines delivered and handles blocking the
// YarnSpinner thread as expected by the game.
type DialogueRunner struct {
	program     *bytecode.Program
	stringTable *yarn.StringTable

	gameState    *GameState
	runState     RunnerState          // runState is manipulated by handler
	CurrNodeName string               // CurrNodeName is the name of the currently running node.
	vm           *yarn.VirtualMachine // vm is the Yarn virtual machine.

	mut *sync.RWMutex

	running bool
}

func NewDialogueRunner() (*DialogueRunner, error) {
	program, st, err := yarn.LoadFilesFS(yarnBin, yarnFile+".yarnc", language.EN_US)
	if err != nil {
		return nil, err
	}
	return &DialogueRunner{
		program:     program,
		stringTable: st,
		runState:    RunnerStopped,
		mut:         &sync.RWMutex{},
	}, nil
}

// Start starts the runner, which blocks the current thread until a fatal error occurs.
func (r *DialogueRunner) Start(vars yarn.MapVariableStorage, handler yarn.DialogueHandler) error {
	r.runState = RunnerRunning

	if vars == nil {
		vars = make(yarn.MapVariableStorage)
	}

	r.vm = &yarn.VirtualMachine{
		Program: r.program,
		Handler: handler,
		Vars:    vars,
	}
	return r.vm.Run("Start")
}

func (r *DialogueRunner) GameState() *GameState {
	r.gameState.Vars = r.vm.Vars.(yarn.MapVariableStorage)
	r.gameState.CurrentNode = r.CurrNodeName
	return r.gameState
}
func (r *DialogueRunner) IsLastLine(line yarn.Line) bool {
	if _, ok := r.stringTable.Table[line.ID]; !ok {
		return false
	}
	fmt.Println("tags", r.stringTable.Table[line.ID].Tags)
	for _, tag := range r.stringTable.Table[line.ID].Tags {
		if tag == "lastline" {
			return true
		}
	}
	return false
}

func (r *DialogueRunner) Render(line yarn.Line) string {
	s, err := r.stringTable.Render(line)
	if err != nil {
		debug.Println("error rendering line", line)
		return "ERROR"
	}
	return s.String()
}
