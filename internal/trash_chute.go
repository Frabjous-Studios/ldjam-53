package internal

type TrashChute struct {
	*BaseSprite

	Contents []Sprite
}

func NewTrashChute() *TrashChute {
	result := &TrashChute{
		BaseSprite: &BaseSprite{X: 282, Y: 225, Img: Resources.images["trash_chute"]},
	}
	return result
}
