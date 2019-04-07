package gozcore

import (
	"fmt"
	"image"
	"time"
)

type protocolv4 struct {
	width, height int
}

func (p protocolv4) GetSize() (width int, height int) {
	return p.width, p.height
}

func (p protocolv4) SetBurnTime(graver *Graver, burn int) error {
	if burn < 1 || burn > 240 {
		return fmt.Errorf("burntime is out of range: [1, 240]")
	}
	graver.Connection.Send([]byte{255, 5, byte(burn), 0})
	time.Sleep(time.Millisecond * 20)
	return nil
}

func (p protocolv4) SetLaserPower(graver *Graver, power int) error {
	if power < 1 || power > 100 {
		return fmt.Errorf("power is out of range: [1, 100]")
	}
	graver.Connection.Send([]byte{255, 13, 0, byte(power)})
	time.Sleep(time.Millisecond * 20)
	return nil
}

func (p protocolv4) Reset(graver *Graver) {
	log.Info("Reset the graver")
	graver.Connection.Send([]byte{255, 4, 1, 0})
}

func (p protocolv4) Engrave(graver *Graver, img image.Image, times int) error {
	var w, h, stride, x, y, gw, gh int

	// Maximum engraving size
	gw, gh = p.GetSize()

	rect := img.Bounds()

	w = rect.Max.X - rect.Min.X
	h = rect.Max.Y - rect.Min.Y

	if w > gw || h > gh {
		return fmt.Errorf("image is too big, maximum allowed is {width: %d, height: %d", gw, gh)
	}

	if w%8 == 0 {
		stride = w / 8
	} else {
		stride = (w / 8) + 1
	}

	/*
		The output keep the same height but compress the width by 8
		The graver expects bits, either 1 for engraving or 0 for not engraving
		An image {width: 24px, height: 4px}, with the half-left black should be
		converted to this:

		111111111111000000000000
		111111111111000000000000
		111111111111000000000000
		111111111111000000000000

		So, we are expecting this bytes:
		ff:f0:00
		ff:f0:00
		ff:f0:00
		ff:f0:00

		data []byte, when initialized, is full of zeroes.

		The following code will go, line by line, on each pixel.
		For each pixel, it will find the corresponding byte
		with the formula: (y * stride) + (x / 8)

		If there any color in this pixel, it will switch the corresponding
		bit, inside the byte, using a modulo 8 and a bit shift.

		128 >> 0 = 10000000
		128 >> 1 = 01000000
		etc...

	*/

	data := make([]byte, stride*h)

	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			// RGBA() returns uint32 values
			cr, cg, cb, _ := img.At(x, y).RGBA()
			// Convert each value to uint8
			r := uint8(cr >> 8)
			g := uint8(cg >> 8)
			b := uint8(cb >> 8)

			i := (y * stride) + (x / 8)

			// If color != FFFFFF
			if r&g&b < 255 {
				data[i] |= 128 >> byte(x%8)
			}
		}
	}

	x = (gw - w) / 2
	y = (gh - h) / 2

	log.Infof("Request the engraving of image {width: %d, height: %d}", w, h)

	// Tell the graver where are the top left coordinates
	graver.Connection.Send([]byte{255, 110, 1, byte(x / 100), byte(x % 100), byte(y / 100), byte(y % 100)})
	log.Debugf("Engraving {x: %d, y: %d}", x, y)
	// And what are the width & height.
	// The width is rounded up because the compression to bits is modulo 8
	graver.Connection.Send([]byte{255, 110, 2, byte(stride * 8 / 100), byte(stride * 8 % 100), byte(h / 100), byte(h % 100)})
	log.Debugf("Engraving {width: %d, height: %d}", stride*8, h)

	time.Sleep(time.Millisecond * 20)

	// Tell the graver we are ready for upload
	graver.Connection.Send([]byte{255, 6, 1, 1})

	// Wait for the graver to acknowledge the upload
	<-graver.Events.Once(EventReadyForUpload)

	time.Sleep(time.Millisecond * 20)

	// Upload
	graver.Connection.Send(data)

	// Wait for the graver to finish
	<-graver.Events.Once(EventEngravingDone)
	log.Infof("Engraving %d/%d done", 1, times)

	// If the engraving has been requested more than once,
	// use the special command {255, 1, 1, 0} to repeat it and wait again
	for i := 1; i < times; i++ {
		graver.Connection.Send([]byte{255, 1, 1, 0})
		<-graver.Events.Once(EventEngravingDone)
		log.Infof("Engraving %d/%d done", i, times)
	}

	return nil
}
