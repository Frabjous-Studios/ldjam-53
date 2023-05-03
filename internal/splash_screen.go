package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

// studiosDelay is the duration before "studios" fades in.
const studiosDelay = time.Second * 2
const outDelay = time.Second * 5
const outTime = 0.5

type LogoScene struct {
	Game *Game

	Logo          *ebiten.Image
	Studios       *ebiten.Image
	LogoShader    *ebiten.Shader
	StudiosShader *ebiten.Shader

	buff *ebiten.Image

	startTime time.Time
}

func NewLogoScene(g *Game) *LogoScene {
	return &LogoScene{
		Game:          g,
		buff:          ebiten.NewImage(g.Width, g.Height),
		Logo:          Resources.GetImage("frabjous.png"),
		Studios:       Resources.GetImage("studios.png"),
		LogoShader:    Resources.GetShader("LogoSplash.kage"),
		StudiosShader: Resources.GetShader("Fade.kage"),
	}
}

func (m *LogoScene) Update() error {
	if m.startTime.IsZero() {
		m.startTime = time.Now()
	}
	if time.Now().Sub(m.startTime) > time.Second*11/2 {
		m.Game.ChangeScene(NewMainMenuScene(m.Game))
	}

	return nil
}

var uniforms = make(map[string]interface{})

func (m *LogoScene) Draw(screen *ebiten.Image) {
	x, y := float64(m.Game.Width-m.Logo.Bounds().Dx())/2.0, float64(m.Game.Height-2*m.Logo.Bounds().Dy())/2.0-40

	dt := time.Now().Sub(m.startTime)
	shopts := ebiten.DrawRectShaderOptions{}
	shopts.Images[0] = m.Logo
	shopts.Uniforms = uniforms
	shopts.Uniforms["Dt"] = dt.Seconds()
	shopts.Uniforms["ScreenPos"] = [2]float32{float32(2 * x), float32(2 * y)}
	shopts.GeoM.Translate(x, y)

	m.buff.DrawRectShader(m.Logo.Bounds().Dx(), m.Logo.Bounds().Dy(), m.LogoShader, &shopts)

	x, y = float64(m.Game.Width-m.Studios.Bounds().Dx())/2.0, float64(m.Game.Height-2*m.Studios.Bounds().Dy())/2.0+10
	shopts = ebiten.DrawRectShaderOptions{}
	shopts.Images[0] = m.Studios
	shopts.Uniforms = uniforms
	shopts.Uniforms["Dt"] = float32((dt - studiosDelay).Seconds())
	shopts.Uniforms["ScreenPos"] = [2]float32{248, 281}
	shopts.GeoM.Translate(x, y)

	m.buff.DrawRectShader(m.Studios.Bounds().Dx(), m.Studios.Bounds().Dy(), m.StudiosShader, &shopts)

	opts := ebiten.DrawImageOptions{}
	if dt >= outDelay {
		t := 1.0 - float32((dt-outDelay).Seconds())/outTime
		opts.ColorScale.Scale(t, t, t, t)
	}
	screen.DrawImage(m.buff, &opts)
}
