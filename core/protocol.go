package gozcore

import "image"

type protocol interface {
	GetSize() (width int, height int)
	SetBurnTime(graver *Graver, burn int) error
	Reset(graver *Graver)
	Engrave(graver *Graver, img image.Image, times int) error
}
