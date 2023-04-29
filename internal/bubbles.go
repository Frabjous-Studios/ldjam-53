package internal

import (
	uiimg "github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"time"
	"unicode/utf8"
)

const CrawlSpeedCPS = 120
const DialogFont = "Munro"
const lineSpacing = 1.15

var (
	fontColor          = h2c("ffffff")
	fontColorHighlight = h2c("ffff00")
)

// bubbleDelay is the min amount of time to show a bubble before moving on to the next dialogue option.
const bubbleDelay = 5 * time.Second

type Bubbles struct {
	txt         *etxt.Renderer
	offscrn     *ebiten.Image // offscreen buffer for text rendering
	bubblePatch *uiimg.NineSlice

	feed *etxt.Feed

	stack        []*Line
	charsShown   int
	scene        *MainScene
	completeTime time.Time
}

func NewBubbles(m *MainScene) *Bubbles {
	result := &Bubbles{
		scene: m,
	}
	result.offscrn = ebiten.NewImage(200, 80)
	result.offscrn.Fill(color.RGBA{R: 0, B: 0, G: 0, A: 0}) // TODO: use acutal ninepatch.
	txt := etxt.NewStdRenderer()
	txt.SetTarget(result.offscrn)
	txt.SetFont(Resources.GetFont(DialogFont))
	txt.SetAlign(etxt.Bottom, etxt.XCenter)
	txt.SetSizePx(fontSize)
	txt.SetColor(fontColor)
	txt.SetLineSpacing(lineSpacing)
	result.txt = txt

	result.bubblePatch = Resources.GetNineSlice("bubble")

	return result
}

func (b *Bubbles) SetLine(str string) {
	b.stack = []*Line{NewLine(str)}
	b.completeTime = time.Time{}
}

func (b *Bubbles) Update() {
	if b.IsDone() {
		if b.completeTime.IsZero() {
			b.completeTime = time.Now()
		}
		if time.Now().Sub(b.completeTime) > bubbleDelay {
			b.scene.speaking.Broadcast()
		}
	}
}

func (b *Bubbles) BeDone() {
	if len(b.stack) == 0 {
		return
	}
	if !b.IsDone() {
		b.stack[0].charsShown = len(b.stack[0].Text)
	} else {
		b.completeTime = time.Now().Add(-bubbleDelay)
	}
}

func (b *Bubbles) IsDone() bool {
	if len(b.stack) == 0 {
		return false
	}
	return b.stack[0].charsShown == len(b.stack[0].Text)
}

var TextBounds = rect(340, 56, 200, 100)

func (b *Bubbles) DrawTo(screen *ebiten.Image) bool {
	if b.Empty() {
		return true
	}
	const padding = 3
	b.txt.SetTarget(b.offscrn)
	feed := b.txt.NewFeed(fixed.P(TextBounds.Min.X, TextBounds.Min.Y))
	// draw text once offscreen to capture rectangles
	for _, l := range b.stack {
		l.Rect = b.print(feed, l, TextBounds)
	}

	//pos := b.Pos()
	for _, line := range b.stack {
		b.bubblePatch.Draw(screen, line.Rect.Dx()+4*padding, line.Rect.Dy()+4*padding, func(opts *ebiten.DrawImageOptions) {
			opts.GeoM.Translate(float64(line.Rect.Min.X-padding), float64(line.Rect.Min.Y-padding))
		})
	}
	b.txt.SetTarget(screen)
	feed = b.txt.NewFeed(fixed.P(TextBounds.Min.X, TextBounds.Min.Y))
	for _, l := range b.stack {
		b.txt.SetColor(fontColor)
		l.Rect = b.print(feed, l, TextBounds)
	}
	return b.IsDone()
}
func (b *Bubbles) Empty() bool {
	return len(b.stack) == 0 || len(b.stack[0].Text) == 0
}

func (b *Bubbles) Bounds() image.Rectangle {
	return rect(340, 12, 200, 72)
}

func (b *Bubbles) Pos() image.Point {
	return image.Pt(340, 12)
}

func (b *Bubbles) SetPos(point image.Point) {
	return // ignore
}

type Line struct {
	Rect        image.Rectangle
	Text        string
	crawlStart  time.Time
	charsShown  int
	highlighted bool
}

// charsToShow yields the number of characters of the currently displaying text to show based on time since the message
// was first shown and the crawl speed. The offset provided is subtracted from the result, and can be used to
func (l *Line) charsToShow() int {
	return int(time.Now().Sub(l.crawlStart).Seconds() * CrawlSpeedCPS)
}

func NewLine(text string) *Line {
	return &Line{
		Text:       text,
		crawlStart: time.Now(),
	}
}

// NewOption returns a line with a crawlStart of zero.
func NewOption(text string) *Line {
	return &Line{Text: text}
}

const fontSize = 16

// modified from etxt examples
func (b *Bubbles) print(feed *etxt.Feed, line *Line, bounds image.Rectangle) image.Rectangle {
	charsToShow := line.charsToShow()
	// helper function
	var getNextWord = func(str string, index int) string {
		start := index
		for index < len(str) {
			codePoint, size := utf8.DecodeRuneInString(str[index:])
			if codePoint <= ' ' {
				return str[start:index]
			}
			index += size
		}
		return str[start:index]
	}
	used := image.Rectangle{}
	used.Min.X = feed.Position.X.Ceil()
	used.Min.Y = feed.Position.Y.Floor() - fontSize // -fontSize enables choice highlighting.
	used.Max = used.Min

	// create Feed and iterate each rune / word
	if feed == nil {
		feed = b.txt.NewFeed(fixed.P(bounds.Min.X, bounds.Min.Y-fontSize)) // -fontSize enables choice highlighting.
	}
	index := 0
	totalChars := 0
	for totalChars < charsToShow && index < len(line.Text) {
		switch line.Text[index] {
		case ' ': // handle spaces with Advance() instead of Draw()
			feed.Advance(' ')
			totalChars++
			index += 1
		case '\n', '\r': // \r\n line breaks *not* handled as single line breaks
			feed.LineBreak()
			totalChars++
			index += 1
		default:
			// get next word and measure it to see if it fits
			word := getNextWord(line.Text, index)
			width := b.txt.SelectionRect(word).Width
			if (feed.Position.X + width).Ceil() > bounds.Max.X {
				feed.LineBreak() // didn't fit, jump to next line before drawing
			}

			// abort if we are going beyond the vertical working area
			if feed.Position.Y.Floor() >= bounds.Max.Y {
				return used
			}
			used.Max.X = max(used.Max.X, (feed.Position.X + width).Ceil())
			used.Max.Y = max(used.Max.Y, feed.Position.Y.Floor())

			// draw the word and increase index
			for _, codePoint := range word {
				feed.Draw(codePoint) // you may want to cut this earlier if the word is too long
				totalChars++
				if totalChars == charsToShow {
					break
				}
			}
			index += len(word)
		}
	}
	line.charsShown = totalChars
	return used
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
