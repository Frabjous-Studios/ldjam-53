package internal

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/colorutil"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kalexmills/asebiten"
	"github.com/solarlune/resound"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
	"image/color"
	_ "image/png"
	"io/fs"
	"math/rand"
	"strings"
)

var bodies = []string{
	"body_armor.png",
	"body_cloak.png",
	"body_jumpsuit.png",
	"body_sleeveless.png",
	"body_sleeveless_alt.png",
	"body_suit.png",
	"body_tshirt.png",
	"body_tshirt_2tone.png",
	"body_tshirt_alt.png",
	"body_tshirt_alt2.png",
	"body_wifebeat.png",
}

var heads = []string{
	"head_3eyeShades.png",
	"head_antlerClops.png",
	"head_apeMojo.png",
	"head_blank.png",
	"head_boomerang.png",
	"head_bulbous.png",
	"head_bunGirl.png",
	"head_cactus.png",
	"head_cat.png",
	"head_eraserGlasses.png",
	"head_gills.png",
	"head_glareGaunt.png",
	"head_gorgeous.png",
	"head_grimFlattop.png",
	"head_insect.png",
	"head_logBirb.png",
	"head_mohawkShades.png",
	"head_pillBot.png",
	"head_ponytails.png",
	"head_psychoClown.png",
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
	players    map[string]*audio.Player
	music      map[string]*audio.InfiniteLoop
	anims      map[string]*asebiten.Animation

	playing *resound.Volume
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

	// anims
	Resources.anims = make(map[string]*asebiten.Animation)
	Resources.anims["shredder"] = Resources.GetAnim("scan_shreder.json")
	Resources.anims["alarm_buttons"] = Resources.GetAnim("alarmbuttons.json")

	// images
	Resources.images = make(map[string]*ebiten.Image)
	Resources.bodies = Resources.loadImages(bodies)
	Resources.heads = Resources.loadImages(heads)

	// load nineslices
	Resources.nineSlices = make(map[string]*image.NineSlice)
	// TODO: use a real image
	Resources.nineSlices["bubble"] = image.NewNineSlice(Resources.GetImage("dialog_9patch2x2cells.png"), [3]int{2, 2, 2}, [3]int{2, 2, 2})

	// load shaders
	Resources.shaders = make(map[string]*ebiten.Shader)
	Resources.shaders["day_night"] = loadShader("DayNight.kage")
	Resources.shaders["LogoSplash.kage"] = loadShader("LogoSplash.kage")
	Resources.shaders["Fade.kage"] = loadShader("Fade.kage")

	// load font faces (again?)
	Resources.faces = make(map[string]*truetype.Font)

	// load audio
	Resources.players = make(map[string]*audio.Player)
	Resources.music = make(map[string]*audio.InfiniteLoop)

	// load images
	Resources.images["frabjous.png"] = Resources.GetImage("frabjous.png")
	Resources.images["studios.png"] = Resources.GetImage("studios.png")
	Resources.images["bill_1"] = Resources.GetImage("bill_1.png")
	Resources.images["bill_5"] = Resources.GetImage("bill_5.png")
	Resources.images["bill_10"] = Resources.GetImage("bill_10.png")
	Resources.images["bill_20"] = Resources.GetImage("bill_20.png")
	Resources.images["bill_100"] = Resources.GetImage("bill_100.png")

	Resources.images["stack_1"] = Resources.GetImage("bill_stack_1.png")
	Resources.images["stack_5"] = Resources.GetImage("bill_stack_5.png")
	Resources.images["stack_10"] = Resources.GetImage("bill_stack_10.png")
	Resources.images["stack_20"] = Resources.GetImage("bill_stack_20.png")
	Resources.images["stack_100"] = Resources.GetImage("bill_stack_100.png")

	Resources.images["coin_1"] = Resources.GetImage("coin_1.png")
	Resources.images["coin_5"] = Resources.GetImage("coin_5.png")
	Resources.images["coin_10"] = Resources.GetImage("coin_10.png")
	Resources.images["coin_25"] = Resources.GetImage("coin_25.png")
	Resources.images["coin_50"] = Resources.GetImage("coin_50.png")

	Resources.images["check_front"] = Resources.GetImage("check_front.png")
	Resources.images["check_back"] = Resources.GetImage("check_back.png")

	Resources.images["deposit_slip_empty"] = Resources.GetImage("deposit_slip.png")
	Resources.images["deposit_slip_deposit"] = Resources.GetImage("deposit_slip_deposit.png")
	Resources.images["deposit_slip_withdrawal"] = Resources.GetImage("deposit_slip_withdrawal.png")

	Resources.images["photo_id"] = Resources.GetImage("photo_id.png")

	Resources.images["drone"] = Resources.GetImage("drone.png")

	Resources.images["call_button_holo"] = Resources.GetImage("call_button_hologram.png")
	Resources.images["call_button"] = Resources.GetImage("call_button.png")
	Resources.images["counter"] = Resources.GetImage("countertop.png")
	Resources.images["terminal"] = Resources.GetImage("terminal_bg.png")
	Resources.images["trash_chute"] = Resources.GetImage("trashchute.png")
	Resources.images["Till"] = Resources.GetImage("placeholer_till.png")

	Resources.images["bg_bg.png"] = Resources.GetImage("bg_bg.png")
	Resources.images["bg_fg.png"] = Resources.GetImage("bg_fg.png")

	Resources.images["junk_1.png"] = Resources.GetImage("junk_1.png")
	Resources.images["junk_2.png"] = Resources.GetImage("junk_2.png")
	Resources.images["junk_3.png"] = Resources.GetImage("junk_3.png")
	Resources.images["junk_4.png"] = Resources.GetImage("junk_4.png")
	Resources.images["junk_5.png"] = Resources.GetImage("junk_5.png")
	Resources.images["junk_6.png"] = Resources.GetImage("junk_6.png")
	Resources.images["junk_7.png"] = Resources.GetImage("junk_7.png")
	Resources.images["junk_8.png"] = Resources.GetImage("junk_8.png")
	Resources.images["junk_9.png"] = Resources.GetImage("junk_9.png")
	Resources.images["junk_10.png"] = Resources.GetImage("junk_10.png")
}

const SampleRate = 44100

//go:embed gamedata/audio
var audioFiles embed.FS

func (r *resources) GetSound(aCtx *audio.Context, filename string) *audio.Player {
	if p, ok := r.players[filename]; ok {
		return p
	}
	b, err := audioFiles.ReadFile(fmt.Sprintf("gamedata/audio/%s", filename))
	if err != nil {
		debug.Printf("error opening audio file: %v", err)
	}

	reader := bytes.NewReader(b)

	stream, err := vorbis.DecodeWithSampleRate(SampleRate, reader)
	if err != nil {
		debug.Printf("error decoding wav file: %v", err)
	}

	p, err := aCtx.NewPlayer(stream)
	if err != nil {
		debug.Printf("error creating a new player: %v", err)
	}
	r.players[filename] = p
	return p
}

func (r *resources) GetRandSound(aCtx *audio.Context, files ...string) *audio.Player {
	f := files[rand.Intn(len(files))]
	return r.GetSound(aCtx, f)
}

func (r *resources) GetMusic(aCtx *audio.Context, file string) *audio.InfiniteLoop {
	if p, ok := r.music[file]; ok {
		return p
	}
	b, err := audioFiles.ReadFile(fmt.Sprintf("gamedata/audio/%s", file))
	if err != nil {
		debug.Printf("error opening audio file: %v", err)
	}

	reader := bytes.NewReader(b)

	stream, err := vorbis.DecodeWithSampleRate(SampleRate, reader)
	if err != nil {
		debug.Printf("error decoding wav file: %v", err)
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	r.music[file] = loop
	return loop
}

func (r *resources) GetAnim(path string) *asebiten.Animation {
	if _, ok := r.anims[path]; !ok {
		a, err := asebiten.LoadAnimation(art, fmt.Sprintf("gamedata/img/%s", path))
		if err != nil {
			debug.Printf("error loading animation: %s: %v", path, err)
		}
		r.anims[path] = a
	}
	return r.anims[path]
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
	if len(path) == 0 {
		panic("eep!")
	}
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

//go:embed gamedata/shader
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

const portraitStartX, portraitStartY = 170, 52

// NewImageColor constructs a new Image that when drawn fills with color c.
func placeholder(c color.Color, w, h int) *ebiten.Image {
	i := ebiten.NewImage(w, h)
	i.Fill(c)
	return i
}

func newPortrait(target *ebiten.Image, body, head string) *Customer {
	b, h := Resources.GetImage(body), Resources.GetImage(head)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(0, 32)
	target.DrawImage(b, opts)
	opts.GeoM.Translate(0, -32)
	target.DrawImage(h, opts)
	return &Customer{
		ImageKey: fmt.Sprintf("%s:%s", body, head),
		BaseSprite: &BaseSprite{
			Img: target,
			X:   portraitStartX,
			Y:   portraitStartY,
		},
	}
}

func newRandPortrait(target *ebiten.Image) *Customer {
	return newPortrait(target, randMapKey(Resources.bodies), randMapKey(Resources.heads))
}

func newSimplePortrait(target *ebiten.Image, head string) *Customer {
	h := Resources.GetImage(head)
	target.DrawImage(h, nil)
	return &Customer{
		ImageKey: head,
		BaseSprite: &BaseSprite{
			Img: target,
			X:   portraitStartX,
			Y:   portraitStartY,
		},
	}
}

func drawRandom[T any](vals []T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	return vals[rand.Intn(len(vals))]
}
