package internal

import (
	"bufio"
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
	"math/rand"
	"strings"
)

var bodies = []string{ // TODO: put them here
	"body_cloak.png",
	"body_jumpsuit.png",
	"body_sleeveless.png",
	"body_tshirt.png",
}

var heads = []string{ // TODO: put them here
	"head_bunGirl.png",
	"head_cactus.png",
	"head_glareGaunt.png",
	"head_grimFlattop.png",
	"head_insect.png",
	"head_mohawkShades.png",
	"head_pillBot.png",
	"head_smileScreen.png",
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
	lists      map[string][]string
}

const FontName = "Munro-2LYe.ttf"

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
	Resources.lists = make(map[string][]string)

	Resources.images = make(map[string]*ebiten.Image)
	Resources.bodies = Resources.loadImages(bodies)
	Resources.heads = Resources.loadImages(heads)

	// load nineslices
	Resources.nineSlices = make(map[string]*image.NineSlice)
	// TODO: use a real image
	Resources.nineSlices["bubble"] = image.NewNineSlice(Resources.GetImage("dialog_9patch2x2cells.png"), [3]int{2, 2, 2}, [3]int{2, 2, 2})

	// load shaders
	Resources.shaders = make(map[string]*ebiten.Shader)

	// load font faces (again?)
	Resources.faces = make(map[string]*truetype.Font)

	Resources.images["bill_1"] = Resources.GetImage("bill_1.png")
	Resources.images["bill_5"] = Resources.GetImage("bill_5.png")
	Resources.images["bill_10"] = Resources.GetImage("bill_10.png")
	Resources.images["bill_20"] = Resources.GetImage("bill_20.png")
	Resources.images["bill_100"] = Resources.GetImage("bill_100.png")

	Resources.images["coin_1"] = placeholder(h2c("ffff00"), 15, 15)
	Resources.images["coin_5"] = placeholder(h2c("dddd00"), 15, 15)
	Resources.images["coin_10"] = placeholder(h2c("cccc00"), 15, 15)
	Resources.images["coin_25"] = placeholder(h2c("bbbb00"), 15, 15)
	Resources.images["coin_50"] = placeholder(h2c("aaaa00"), 15, 15)

	Resources.images["check_front"] = Resources.GetImage("check_front.png")
	Resources.images["check_back"] = Resources.GetImage("check_back.png")

	Resources.images["deposit_slip_empty"] = Resources.GetImage("deposit_slip.png")
	Resources.images["deposit_slip_deposit"] = Resources.GetImage("deposit_slip_deposit.png")
	Resources.images["deposit_slip_withdrawal"] = Resources.GetImage("deposit_slip_withdrawal.png")

	Resources.images["photo_id"] = Resources.GetImage("photo_id.png")

	Resources.images["counter"] = placeholder(h2c("ff0000"), 208, 88)
	Resources.images["Till"] = Resources.GetImage("placeholer_till.png")

	Resources.images["bg_bg.png"] = Resources.GetImage("bg_bg.png")
	Resources.images["bg_fg.png"] = Resources.GetImage("bg_fg.png")
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

func (r *resources) RandomScriptFont() *etxt.Font {
	var options = []string{"Honey Script Light", "Roustel Regular", "Thesignature"}
	chosen := options[rand.Intn(3)]
	return r.GetFont(chosen)
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

//go:embed gamedata/*.txt
var text embed.FS

func (r *resources) GetList(filename string) []string {
	if result, ok := r.lists[filename]; ok {
		return result
	}
	f, err := text.Open(fmt.Sprintf("gamedata/%s", filename))
	if err != nil {
		debug.Printf("error opening text file %s: %v", filename, err)
	}
	s := bufio.NewScanner(f)
	var result []string
	for s.Scan() {
		result = append(result, s.Text())
	}
	r.lists[filename] = result
	return result
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
	opts.GeoM.Translate(0, 32)
	target.DrawImage(b, opts)
	opts.GeoM.Translate(0, -32)
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

func drawRandom[T any](vals []T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	return vals[rand.Intn(len(vals))]
}
