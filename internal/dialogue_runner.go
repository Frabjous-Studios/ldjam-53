package internal

import (
	"embed"
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/DrJosh9000/yarn/bytecode"
	"github.com/Frabjous-Studios/bankwave/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/razor-1/localizer-cldr/resources/language"
	"strings"
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
	runState     RunnerState // runState is manipulated by handler
	CurrNodeName string      // CurrNodeName is the name of the currently running node.

	mut *sync.RWMutex
	vm  *yarn.VirtualMachine // vm is the Yarn virtual machine.

	portraitImg *ebiten.Image
	customer    *Customer

	dialogueLines   chan string
	dialogueOptions chan int

	fullName, firstName, lastName string

	running bool
}

func NewDialogueRunner(vars yarn.MapVariableStorage, handler yarn.DialogueHandler) (*DialogueRunner, error) {
	program, st, err := yarn.LoadFilesFS(yarnBin, yarnFile+".yarnc", language.EN_US)
	if err != nil {
		return nil, err
	}
	r := &DialogueRunner{
		program:     program,
		stringTable: st,
		runState:    RunnerStopped,
		mut:         &sync.RWMutex{},
	}
	r.vm = &yarn.VirtualMachine{
		Program: r.program,
		Handler: handler,
		Vars:    vars,
	}
	return r, nil
}

const (
	VarFullName      = "$char_full_name"
	VarFirstName     = "$char_first_name"
	VarLastName      = "$char_last_name"
	VarSlipAmt       = "$slip_amount"
	VarAccountNumber = "$account_number"
)

// DoNode starts the runner, which blocks the current thread until a fatal error occurs.
func (r *DialogueRunner) DoNode(name string) error {
	// TODO: this should generate a new customer picture and identity before starting *anything* so we can avoid race conditions.
	defer func() {
		r.runState = RunnerStopped
	}()
	debug.Println("doing node; clearing DialogueRunner customer")
	r.CurrNodeName = name
	r.running = true
	r.customer = nil
	r.runState = RunnerRunning

	return r.vm.Run(name)
}

func (r *DialogueRunner) RandomName() string {
	f, l := drawRandom(Resources.GetList("first_names.txt")), drawRandom(Resources.GetList("last_names.txt"))

	r.mut.Lock()
	defer r.mut.Unlock()
	fullName := fmt.Sprintf("%s %s", f, l)
	r.fullName = f
	r.lastName = l
	r.fullName = fullName
	//r.vm.Vars.SetValue(VarFirstName, f) TODO: use once we can edit content again
	//r.vm.Vars.SetValue(VarLastName, f)
	//r.vm.Vars.SetValue(VarFullName, fullName)
	return fullName
}

func (r *DialogueRunner) FullName() string {
	return r.fullName // TODO: r.getString(VarFullName)
}

func (r *DialogueRunner) FirstName() string {
	return r.firstName // TODO: r.getString(VarFirstName)
}

func (r *DialogueRunner) LastName() string {
	return r.lastName // TODO: r.getString(VarLastName)
}

// SetDepositSlip sets variables associated with the generated deposit slip.
func (r *DialogueRunner) SetDepositSlip(slip *DepositSlip) {
	//r.SetDepositAmt(slip.Value)  TODO: use after jam, when we can update content again
	//r.SetAccountNumber(slip.AcctNum)
	if r.customer != nil {
		r.customer.DepositSlip = slip
	}
}

func (r *DialogueRunner) SetDepositAmt(val int) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.vm.Vars.SetValue(VarSlipAmt, fmt.Sprintf("%d.%02d", val/100, val%100))
}
func (r *DialogueRunner) SetAccountNumber(val int) {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.vm.Vars.SetValue(VarAccountNumber, val)
}

// CustomerIntent gets the intent set for this node.
func (r *DialogueRunner) CustomerIntent(currNode string) Intent {
	node, ok := r.vm.Program.Nodes[currNode]
	if !ok {
		debug.Printf("could not find node %v when looking for intent", r.CurrNodeName)
		return ""
	}
	for _, h := range node.Headers {
		if h.Key == "intent" {
			switch h.Value {
			case IntentCashCheck:
				return IntentCashCheck
			case IntentDeposit:
				return IntentDeposit
			case IntentDepositCheck:
				return IntentDepositCheck
			case IntentWithdraw:
				return IntentWithdraw
			default:
				debug.Printf("Unknown intent '%s' in node %s", h.Value, r.CurrNodeName)
			}
		}
	}
	return ""
}

func (r *DialogueRunner) PortraitID(nodeName string) string {
	node, ok := r.vm.Program.Nodes[nodeName]
	if !ok {
		return ""
	}
	return portrait(node)
}

func (r *DialogueRunner) Customer(nodeID string) (p *Customer) {
	defer func() {
		// TODO: ths defer makes this func a bit weird.
		if p != nil {
			r.customer = p
			p.CustomerIntent = r.CustomerIntent(nodeID)
			p.CustomerName = r.RandomName()
			p.IsRude = strings.Contains(strings.ToLower(nodeID), "rude")
		}
	}()
	portraitID := r.PortraitID(nodeID)
	if portraitID == "" {
		debug.Println("missing portraitID in node", nodeID)
		portraitID = "random"
	}
	if portraitID == "random" {
		return newRandPortrait()
	}
	toks := strings.Split(portraitID, ":")
	if len(toks) == 1 {
		return newSimplePortrait(toks[0])
	}
	if len(toks) != 2 {
		debug.Printf("malformed customer portraitID! using random: %v", portraitID)
		return newRandPortrait()
	}
	head, body := toks[0], toks[1]
	return newPortrait(body, head)
}

func (r *DialogueRunner) IsLastLine(line yarn.Line) bool {
	if _, ok := r.stringTable.Table[line.ID]; !ok {
		return false
	}
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

func portrait(node *bytecode.Node) string {
	if node == nil {
		return ""
	}
	for _, h := range node.Headers {
		if h.Key == "portrait" {
			return h.Value
		}
	}
	return ""
}

func (r *DialogueRunner) getString(varName string) string {
	r.mut.RLock()
	defer r.mut.RUnlock()

	v, ok := r.vm.Vars.GetValue(VarFullName)
	if !ok {
		return ""
	}
	return v.(string)
}
