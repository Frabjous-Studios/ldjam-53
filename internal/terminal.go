package internal

import (
	"fmt"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"strings"
	"time"
)

type Terminal struct {
	*BaseSprite
	till *Till

	ui          *ebitenui.UI
	rootWindow  *widget.Window
	operational bool

	bg *ebiten.Image
}

const repeatInterval = 300 * time.Millisecond

func logHandler(args *widget.TextInputChangedEventArgs) {
	fmt.Println("changed:", args.InputText)
}

func onlyNumbers(newText string) (bool, *string) {
	fmt.Println("onlyNumbersCalled", newText)
	var result strings.Builder
	for _, r := range []rune(newText)[:6] {
		if '0' <= r && r <= '9' {
			result.WriteRune(r)
		}
	}
	s := result.String()
	return true, &s
}

func NewTerminal(till *Till) *Terminal {
	result := &Terminal{
		till: till,
		BaseSprite: &BaseSprite{
			X: 0,
			Y: 72,
		},
		operational: true, // TODO: turn this OFF for the first day!
	}
	result.bg = Resources.images["terminal"]
	result.Img = ebiten.NewImage(result.bg.Bounds().Dx(), result.bg.Bounds().Dy())
	fmt.Println(result.bg, result.Img)

	result.ui = &ebitenui.UI{
		Container: widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout())),
	}

	// construct a new container that serves as the rootWindow of the UI hierarchy
	rootContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(transparent)),

		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	face := Resources.GetFace("MunroSmall-wPZw.ttf", 12)

	// Create the first tab
	// A TabBookTab is a labelled container. The text here is what will show up in the tab button
	//tabRed := widget.NewTabBookTab("Accounts",
	//	widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(transparent)),
	//	widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	//)
	//
	//
	//redBtn := widget.NewText(
	//	widget.TextOpts.Text("Accounts Tab", face, color.White),
	//	widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
	//		HorizontalPosition: widget.AnchorLayoutPositionCenter,
	//		VerticalPosition:   widget.AnchorLayoutPositionCenter,
	//	})),
	//)
	//tabRed.AddChild(redBtn)
	ledgerTab := widget.NewTabBookTab("Till",
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(transparent)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	ledgerContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(transparent)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
		)),
	)
	ledgerContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("Till Ledger", face, color.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	))
	ledgerTab.AddChild(ledgerContainer)

	input := widget.NewTextInput(
		widget.TextInputOpts.RepeatInterval(repeatInterval),
		widget.TextInputOpts.Face(face),
		widget.TextInputOpts.Color(textInputColor()),
		widget.TextInputOpts.ChangedHandler(logHandler),
		widget.TextInputOpts.Image(textInputImage()),
		widget.TextInputOpts.Validation(onlyNumbers),
		widget.TextInputOpts.ClearOnSubmit(false),
		widget.TextInputOpts.IgnoreEmptySubmit(true),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(face, 2),
		),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
	)
	ledgerContainer.AddChild(input)

	tabBook := widget.NewTabBook(
		widget.TabBookOpts.TabButtonImage(buttonImage()),
		widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{Idle: color.White}),
		widget.TabBookOpts.TabButtonSpacing(5),
		widget.TabBookOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal:  true,
				StretchVertical:    true,
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			),
		),
		widget.TabBookOpts.TabButtonOpts(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(50, 0)),
		),
		widget.TabBookOpts.Tabs(ledgerTab),
	)

	rootContainer.AddChild(tabBook)

	result.rootWindow = widget.NewWindow(
		widget.WindowOpts.Contents(rootContainer),
	)
	result.rootWindow.SetLocation(rect(3*ScaleFactor, 76*ScaleFactor, 105*ScaleFactor, 68*ScaleFactor))

	result.ui.AddWindow(result.rootWindow)

	return result
}

func textInputColor() *widget.TextInputColor {
	return &widget.TextInputColor{
		Idle:          color.White,
		Disabled:      color.White,
		Caret:         color.White,
		DisabledCaret: color.White,
	}
}

var transparent = color.RGBA{0, 0, 0, 0}

func buttonImage() *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle:         image.NewNineSliceColor(transparent),
		Hover:        image.NewNineSliceColor(transparent),
		Pressed:      image.NewNineSliceColor(transparent),
		PressedHover: image.NewNineSliceColor(transparent),
		Disabled:     image.NewNineSliceColor(transparent),
	}
}

func textInputImage() *widget.TextInputImage {
	return &widget.TextInputImage{
		Idle:     image.NewNineSliceColor(transparent),
		Disabled: image.NewNineSliceColor(transparent),
	}
}

func (t *Terminal) DrawTo(screen *ebiten.Image) {
	t.Img.Clear()
	// draw background
	opts := &ebiten.DrawImageOptions{}
	t.Img.DrawImage(t.bg, opts)

	if !t.operational {
		t.BaseSprite.DrawTo(screen)
		return
	}

	// draw ui
	t.BaseSprite.DrawTo(screen)
	t.ui.Draw(screen)
}

func (t *Terminal) Update() {
	t.ui.Update()
}
