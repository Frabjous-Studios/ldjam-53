package internal

import (
	"fmt"
	"github.com/Frabjous-Studios/bankwave/internal/debug"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/colorutil"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"os"
)

type MainMenuScene struct {
	Game    *Game
	ui      *ebitenui.UI
	buttons []*widget.Button

	bg *ebiten.Image

	selected int
}

var newKeys []ebiten.Key
var heldKeys []ebiten.Key

const MainMenuMusic = "Main_Menu.ogg"

func NewMainMenuScene(game *Game) *MainMenuScene {
	result := &MainMenuScene{
		Game:     game,
		selected: -1,
		bg:       Resources.GetImage("menu_bg"),
	}
	result.ui, _ = result.createMenuUI()
	game.PlayMusic(MainMenuMusic)
	return result
}

func (m *MainMenuScene) Update() error {
	m.ui.Update()
	newKeys = inpututil.AppendJustPressedKeys(newKeys)

	for _, key := range newKeys {
		if key == ebiten.KeyUp || key == ebiten.KeyW {
			m.up()
		}
		if key == ebiten.KeyDown || key == ebiten.KeyS {
			m.down()
		}
	}
	m.updateButtons()

	return nil
}

func (m *MainMenuScene) up() {
	if m.selected == -1 {
		m.selected = 0
	} else {
		m.selected = m.selected - 1
		if m.selected < 0 {
			m.selected = 0
		}
	}
}

func (m *MainMenuScene) down() {
	if m.selected == -1 {
		m.selected = 0
	} else {
		m.selected = m.selected + 1
		if m.selected > len(m.buttons)-1 {
			m.selected = len(m.buttons) - 1
		}
	}
}

func (m *MainMenuScene) updateButtons() {
	for idx, btn := range m.buttons {
		if idx == m.selected {
			btn.Focus(true)
		} else {
			btn.Focus(false)
		}
	}
}

func (m *MainMenuScene) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(ScaleFactor, ScaleFactor)
	screen.DrawImage(m.bg, opts)

	m.ui.Draw(screen)
}

func (m *MainMenuScene) createMenuUI() (*ebitenui.UI, error) {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
	)
	btnContainer := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	buttons := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{false}, []bool{false, false, false, false}),
			widget.GridLayoutOpts.Padding(widget.Insets{Top: 20, Bottom: 20}),
			widget.GridLayoutOpts.Spacing(0, 20),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	)

	m.buttons = []*widget.Button{
		m.button("New Game", newGame),
		m.button("Credits", showCredits),
		m.button("Exit", exitGame),
	}
	for _, b := range m.buttons {
		buttons.AddChild(b)
	}

	btnContainer.AddChild(buttons)
	rootContainer.AddChild(btnContainer)
	return &ebitenui.UI{
		Container: rootContainer,
	}, nil
}

func newGame(g *Game) {
	debug.Println("New game clicked")
	g.ChangeScene(NewMainScene(g))
}

func showCredits(g *Game) {
	debug.Println("Credits clicked")
	c, err := NewCreditsScene(g)
	if err != nil {
		debug.Printf("error moving to credits screen: %v", err)
		panic(err)
	}
	g.ChangeScene(c)
}

func exitGame(_ *Game) {
	debug.Println("Exit game clicked")
	os.Exit(0)
}

var hover = color.RGBA{150, 150, 150, 70}
var transparent = color.RGBA{0, 0, 0, 0}

func (m *MainMenuScene) button(text string, onClick func(g *Game)) *widget.Button {
	c := widget.ButtonTextColor{
		//50 49 59
		Idle:     color.RGBA{50, 49, 59, 255},
		Disabled: color.RGBA{50, 49, 59, 255},
	}
	return widget.NewButton(
		widget.ButtonOpts.Text(text, Resources.GetFace(FontName, 32), &c),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         image.NewNineSliceColor(transparent),
			Hover:        image.NewNineSliceColor(hover),
			Pressed:      image.NewNineSliceColor(transparent),
			PressedHover: image.NewNineSliceColor(transparent),
			Disabled:     image.NewNineSliceColor(transparent),
		}),
		widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
			onClick(m.Game)
		}),
		widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			m.selected = -1
			for _, btn := range m.buttons {
				btn.Focus(false)
			}
		}),
		widget.ButtonOpts.TextPadding(widget.Insets{Left: 20, Right: 20}),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})))
}

// hexColor takes a hex string as input and returns a color or panics
func hexColor(hexStr string) color.Color {
	c, err := colorutil.HexToColor(hexStr)
	if err != nil {
		panic(fmt.Errorf("unparseable hexColor '%s': %w", hexStr, err))
	}
	return c
}
