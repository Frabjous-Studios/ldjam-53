package internal

import (
	"fmt"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math/rand"
	"time"
)

type Silhouettes struct {
	*BaseSprite

	particles []*Silhouette

	flyImages  []*ebiten.Image
	walkImages []*ebiten.Image

	density      int
	debounceTime time.Time
}

var WindowBounds = rect(57, 20, 215, 68)

type Silhouette struct {
	*BaseSprite
	VelX, VelY float64
	Flying     bool
	Seen       bool
	Active     bool
}

func (s *Silhouette) DrawTo(screen *ebiten.Image) {
	if s.Img == nil {
		debug.Println("image for sprite was nil at point:", s.X, s.Y)
		return
	}
	opt := &ebiten.DrawImageOptions{}
	if s.VelX < 0 {
		opt.GeoM.Scale(-1, 1) // horizontal flip
		opt.GeoM.Translate(float64(s.Img.Bounds().Dx()), 0)
	}
	opt.GeoM.Translate(float64(s.X), float64(s.Y))
	opt.GeoM.Scale(ScaleFactor, ScaleFactor)
	screen.DrawImage(s.Img, opt)
}

func NewSilhouettes() *Silhouettes {
	result := &Silhouettes{
		flyImages: []*ebiten.Image{
			Resources.GetImage("silhouette_flying_1.png"),
			Resources.GetImage("silhouette_flying_2.png"),
		},
		walkImages: []*ebiten.Image{
			Resources.GetImage("silhouette_walking_1.png"),
			Resources.GetImage("silhouette_walking_2.png"),
			Resources.GetImage("silhouette_walking_3.png"),
			Resources.GetImage("silhouette_walking_4.png"),
			Resources.GetImage("silhouette_walking_5.png"),
			Resources.GetImage("silhouette_walking_6.png"),
			Resources.GetImage("silhouette_walking_7.png"),
		},
		BaseSprite: &BaseSprite{},
	}

	result.particles = make([]*Silhouette, 50)
	for i := 0; i < 50; i++ {
		result.particles[i] = &Silhouette{
			BaseSprite: &BaseSprite{},
		}
		result.reset(result.particles[i])
		result.particles[i].Active = false
	}
	for i := 0; i < MinDensity; i++ {
		result.reset(result.particles[i])
	}
	return result
}

func (s *Silhouettes) DrawTo(screen *ebiten.Image) {
	for _, p := range s.particles {
		p.DrawTo(screen)
	}
}

func (s *Silhouettes) Bounds() image.Rectangle {
	return WindowBounds
}

const MinDensity = 5
const MaxDensity = 20

func (s *Silhouettes) Update() {
	activeCount := 0
	for _, p := range s.particles {
		if p.Active {
			p.MoveX(p.VelX / TPS)
			p.MoveY(p.VelY / TPS)
		}
		if p.Bounds().Overlaps(WindowBounds) {
			p.Seen = true
		} else if p.Seen {
			s.reset(p)
		}
		if p.Active {
			activeCount++
		}
	}
	if activeCount < MinDensity {
		activateCount := MinDensity - activeCount
		fmt.Println("activating!")
		for i := 0; i < activateCount; i++ {
			for _, p := range s.particles {
				if !p.Active {
					s.reset(p)
					activeCount++
					break
				}
			}
		}
	}
	s.density = activeCount
}

const FlyChance = 0.2
const MinVelocity = 25
const MaxVelocity = 50

func (s *Silhouettes) reset(p *Silhouette) {
	if s.density >= MaxDensity {
		return
	}
	if rand.Float32() < FlyChance {
		p.Img = randSlice(s.flyImages)
		p.Flying = true
		p.Y = rand.Intn(38) + s.Y + 10
	} else {
		p.Img = randSlice(s.walkImages)
		p.Flying = false
		p.Y = s.Bounds().Max.Y - p.Img.Bounds().Dy()
	}

	if rand.Float32() < 0.5 {
		p.X = 0
		p.VelX = (MaxVelocity-MinVelocity)*rand.Float64() + MinVelocity
	} else {
		p.X = 320
		p.VelX = -((MaxVelocity-MinVelocity)*rand.Float64() + MinVelocity)
	}
	p.Seen = false
	p.Active = true
}
