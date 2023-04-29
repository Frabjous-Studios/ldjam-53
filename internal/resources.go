package internal

import (
	"embed"
	"fmt"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/colorutil"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
	"image/color"
	_ "image/png"
	"io/fs"
	"strings"
)

var bodies = []string{ // TODO: put them here
	"cloak.png",
}

var heads = []string{ // TODO: put them here
	"head.png",
	"head_1.png",
	"head_2.png",
}

// Resources makes all multimedia resources for the game available.
var Resources = resources{}

type resources struct {
	fontLib    *etxt.FontLibrary
	faces      map[string]*truetype.Font
	nineSlices map[string]*image.NineSlice
	images     map[string]*ebiten.Image
	shaders    map[string]*ebiten.Shader
	heads      map[string]*ebiten.Image
	bodies     map[string]*ebiten.Image
}

const LunchtimeFont = "lunchds.ttf"

//go:embed gamedata/fonts
var fonts embed.FS

func init() {
	// load fonts
	fontLib := etxt.NewFontLibrary()
	_, _, err := fontLib.ParseEmbedDirFonts("gamedata/fonts", fonts)
	if err != nil {
		panic(err)
	}
	var fonts []string
	_ = fontLib.EachFont(func(s string, font *etxt.Font) error {
		fonts = append(fonts, s)
		return nil
	})
	debug.Println("fonts available:", strings.Join(fonts, ","))
	Resources.fontLib = fontLib

	Resources.images = make(map[string]*ebiten.Image)
	Resources.bodies = Resources.loadImages(bodies)
	Resources.heads = Resources.loadImages(heads)

	// load nineslices
	Resources.nineSlices = make(map[string]*image.NineSlice)
	// TODO: use a real image
	Resources.nineSlices["bubble"] = image.NewNineSliceColor(color.RGBA{R: 50, B: 50, G: 50, A: 50})

	// load shaders
	Resources.shaders = make(map[string]*ebiten.Shader)

	// load font faces (again?)
	Resources.faces = make(map[string]*truetype.Font)

	Resources.images["bill_1"] = placeholder(h2c("00ff00"), 18, 43)
	Resources.images["bill_5"] = placeholder(h2c("00aa00"), 18, 43)
	Resources.images["bill_10"] = placeholder(h2c("00cc00"), 18, 43)
	Resources.images["bill_20"] = placeholder(h2c("00dd00"), 18, 43)
	Resources.images["bill_100"] = placeholder(h2c("00ee00"), 18, 43)

	Resources.images["coin_1"] = placeholder(h2c("ffff00"), 15, 15)
	Resources.images["coin_5"] = placeholder(h2c("dddd00"), 15, 15)
	Resources.images["coin_10"] = placeholder(h2c("cccc00"), 15, 15)
	Resources.images["coin_25"] = placeholder(h2c("bbbb00"), 15, 15)
	Resources.images["coin_50"] = placeholder(h2c("aaaa00"), 15, 15)

	Resources.images["counter"] = placeholder(h2c("ff0000"), 208, 88)
	Resources.images["Till"] = placeholder(h2c("0000ff"), 112, 68)
}

// GetFont returns the loaded font if it exists, nil otherwise.
func (r *resources) GetFont(fontName string) *etxt.Font {
	return r.fontLib.GetFont(fontName)
}

// GetNineSlice returns the loaded nineslice if it exists, nil otherwise.
func (r *resources) GetNineSlice(id string) *image.NineSlice {
	return r.nineSlices[id]
}

// loadImages loads a separate map of special images, given the provided paths.
func (r *resources) loadImages(paths []string) map[string]*ebiten.Image {
	result := make(map[string]*ebiten.Image)
	for _, path := range paths {
		result[path] = r.GetImage(path)
	}
	return result
}

//go:embed gamedata/img
var art embed.FS

func (r *resources) GetImage(path string) *ebiten.Image {
	if r.images == nil {
		r.images = make(map[string]*ebiten.Image)
	}
	if _, ok := r.images[path]; !ok {
		bgPath := fmt.Sprintf("gamedata/img/%s", path)
		img, _, err := ebitenutil.NewImageFromFileSystem(art, bgPath)
		if err != nil {
			debug.Printf("failed to load image: %s: %v", bgPath, err)
			return nil
		}
		r.images[path] = img
	}
	return r.images[path]
}

const dpi = 72

// GetFace returns a new font face with the provided FontID and size.
func (r *resources) GetFace(path string, size float64) font.Face {
	if _, ok := r.faces[path]; !ok {
		f, err := loadFont(path)
		if err != nil {
			debug.Printf("error load face %s", path)
			return nil
		}
		r.faces[path] = f
	}

	return truetype.NewFace(r.faces[path], &truetype.Options{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// GetShader retrieves the shader with the provided ID
func (r *resources) GetShader(path string) *ebiten.Shader {
	return r.shaders[path]
}

func loadFont(filename string) (*truetype.Font, error) {
	fontData, err := fs.ReadFile(fonts, "gamedata/fonts/"+filename)
	if err != nil {
		return nil, fmt.Errorf("loadFace: %w", err)
	}
	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}
	return ttfFont, nil
}

// TODO: enable //go:embed gamedata/shader
var shader embed.FS

// loadShader loads and compiles the provided shader.
func loadShader(name string) *ebiten.Shader {
	bytes, err := fs.ReadFile(shader, "gamedata/shader/"+name)
	if err != nil {
		panic(fmt.Errorf("while loading shader: %w", err))
	}
	shader, err := ebiten.NewShader(bytes)
	if err != nil {
		panic(fmt.Errorf("while compiling shader: %w", err))
	}
	return shader
}

func h2c(hex string) color.Color {
	c, _ := colorutil.HexToColor(hex)
	return c
}

// NewImageColor constructs a new Image that when drawn fills with color c.
func placeholder(c color.Color, w, h int) *ebiten.Image {
	i := ebiten.NewImage(w, h)
	i.Fill(c)
	return i
}

func newPortrait(target *ebiten.Image, body, head string) Sprite {
	b, h := Resources.GetImage(body), Resources.GetImage(head)
	opts := &ebiten.DrawImageOptions{}
	target.DrawImage(b, opts)
	target.DrawImage(h, opts)
	return &Portrait{
		BaseSprite: &BaseSprite{
			Img: target,
			X:   170,
			Y:   52,
		},
	}
}

func newRandPortrait(target *ebiten.Image) Sprite {
	return newPortrait(target, randMapKey(Resources.bodies), randMapKey(Resources.heads))
}
