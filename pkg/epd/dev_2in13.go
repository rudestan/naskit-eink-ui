package epd

import (
	"errors"
	"image"
)

var (
	dev2in13LutFullUpdate = []byte{
		0x40, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x80, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x0A, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,	0x00, 0x00,
		0x15, 0x41, 0xA8, 0x32, 0x50, 0x2C, 0x0B,
	}

	dev2in13LutPartialUpdate = []byte{
		0x40, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x80, 0x00,	0x00, 0x00,	0x00, 0x00, 0x00, 0x00,	0x00, 0x00,
		0x40, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x80, 0x00, 0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00, 0x00, 0x00,
		0x0A, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x00, 0x00,	0x00, 0x00, 0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,	0x00, 0x00,
		0x00, 0x00,	0x00, 0x00,	0x00, 0x00, 0x00, 0x00,	0x00, 0x00,
		0x15, 0x41,	0xA8, 0x32,	0x50, 0x2C, 0x0B,
	}

	dev2in13Width  = 122
	dev2in13Height = 250
)

const (
	BgColorWhite = 0xff
	BgColorBlack = 0x00
)

type device interface {
	initBoard() error
	init(partial bool) error
	clear(bgColor byte)
	display(img image.RGBA)
	sleep()
	reset()
}

type dev2in13 struct {
	board *board
	partial bool
}

func newDev2in13(board *board) device {
	return &dev2in13{board, false}
}

func (d *dev2in13) initBoard() error {
	if !d.board.isConnected() {
		if err := d.board.init(); err != nil {
			return err
		}
	}

	return nil
}

func (d *dev2in13) init(partial bool) error {
	if !d.board.isConnected() {
		return errors.New("board is not inited")
	}

	d.partial = partial

	if partial {
		d.initLutPartial()
	} else {
		d.initLutFull()
	}

	return nil
}

func (d *dev2in13) clear(bgColor byte) {
	d.setWindow(0, 0, uint(dev2in13Width-1), uint(dev2in13Height-1))
	for j := 0; j < dev2in13Height; j++ {
		d.setCursor(0, uint(j))
		d.sendCmd(0x24)
		for i := 0; i <= dev2in13Width/8; i++ {
			d.sendData(bgColor)
		}
	}
	d.turnOnDisplay()
}

func (d *dev2in13) display(img image.RGBA) {
	d.setWindow(0, 0, uint(dev2in13Width-1), uint(dev2in13Height-1))

	vertical := d.isVertical(img)

	for y := 0; y < dev2in13Height; y++ {
		d.setCursor(0, uint(y))
		d.sendCmd(0x24)
		yy := y

		if !vertical {
			yy = dev2in13Height - y - 1
		}

		for x := 0; x <= dev2in13Width/8; x++ {
			d.sendData(getImageByte(yy, x, img, vertical))
		}
	}

	if d.partial {
		d.turnOnDisplayPartial()
	} else {
		d.turnOnDisplay()
	}
}

func (d *dev2in13) isVertical(img image.RGBA) bool {
	return img.Rect.Dx() == dev2in13Width && img.Rect.Dy() == dev2in13Height
}

func (d *dev2in13) sleep() {
	d.sendCmd(0x10)
	d.sendData(0x01)
	d.board.cleanup()
}

func (d *dev2in13) initLutPartial()  {
	d.reset()
	d.board.readBUSY()
	d.sendCmd(0x32)
	d.sendData(dev2in13LutPartialUpdate...)
	d.sendCmd(0x37)
	d.sendData(0x00)
	d.sendData(0x00)
	d.sendData(0x00)
	d.sendData(0x00)
	d.sendData(0x00)
	d.sendData(0x40)
	d.sendData(0x00)

	d.sendCmd(0x22)
	d.sendData(0xC0)
	d.sendCmd(0x20)
	d.board.readBUSY()
}

func (d *dev2in13) initLutFull() {
	d.reset()
	d.board.readBUSY()
	d.sendCmd(0x12)
	d.board.readBUSY()

	d.sendCmd(0x74)
	d.sendData(0x54)
	d.sendCmd(0x7E)
	d.sendData(0x3B)

	d.sendCmd(0x01)
	d.sendData(0xF9)
	d.sendData(0x00)
	d.sendData(0x00)

	d.sendCmd(0x11)
	d.sendData(0x01)

	d.sendCmd(0x44)
	d.sendData(0x00)
	d.sendData(0x0F)

	d.sendCmd(0x45)
	d.sendData(0xF9)
	d.sendData(0x00)
	d.sendData(0x00)
	d.sendData(0x00)

	d.sendCmd(0x3C)
	d.sendData(0x03)

	d.sendCmd(0x2C)
	d.sendData(0x50)

	d.sendCmd(0x03)
	d.sendData(dev2in13LutFullUpdate[100])

	d.sendCmd(0x04)
	d.sendData(dev2in13LutFullUpdate[101])
	d.sendData(dev2in13LutFullUpdate[102])
	d.sendData(dev2in13LutFullUpdate[103])

	d.sendCmd(0x3A)
	d.sendData(dev2in13LutFullUpdate[105])
	d.sendCmd(0x3B)
	d.sendData(dev2in13LutFullUpdate[106])

	d.sendCmd(0x32)
	d.sendData(dev2in13LutFullUpdate...)

	d.sendCmd(0x4E)
	d.sendData(0x00)
	d.sendCmd(0x4F)
	d.sendData(0xF9)
	d.sendData(0x00)
	d.board.readBUSY()
}

func (d *dev2in13) reset() {
	d.board.writeRST(true)
	d.board.delayms(200)
	d.board.writeRST(false)
	d.board.delayms(10)
	d.board.writeRST(true)
	d.board.delayms(200)
}

func (d *dev2in13) sendCmd(b byte) {
	d.board.writeDC(false)
	d.board.writeCS(false)
	d.board.writeSPI(b)
	d.board.writeCS(true)
}

func (d *dev2in13) sendData(b ...byte) {
	d.board.writeDC(true)
	d.board.writeCS(false)
	d.board.writeSPI(b...)
	d.board.writeCS(true)
}

func (d *dev2in13) turnOnDisplay() {
	d.sendCmd(0x22)
	d.sendData(0xc7)
	d.sendCmd(0x20)
	d.sendCmd(0xff)
	d.board.readBUSY()
}

func (d *dev2in13) turnOnDisplayPartial() {
	d.sendCmd(0x22)
	d.sendData(0x0c)
	d.sendCmd(0x20)
	d.sendCmd(0xff)
	d.board.readBUSY()
}

func (d *dev2in13) setWindow(xStart, yStart, xEnd, yEnd uint) {
	d.sendCmd(0x44)
	// x point must be the multiple of 8 or the last 3 bits will be ignored
	d.sendData(byte((xStart >> 3) & 0xff))
	d.sendData(byte((xEnd >> 3) & 0xff))
	d.sendCmd(0x45)
	d.sendData(byte(yStart & 0xff))
	d.sendData(byte((yStart >> 8) & 0xff))
	d.sendData(byte(yEnd & 0xff))
	d.sendData(byte((yEnd >> 8) & 0xff))
}

func (d *dev2in13) setCursor(x, y uint) {
	d.sendCmd(0x4e)
	// x point must be the multiple of 8 or the last 3 bits will be ignored
	d.sendData(byte((x >> 3) & 0xff))
	d.sendCmd(0x4f)
	d.sendData(byte(y & 0xff))
	d.sendData(byte((y >> 8) & 0xff))
	d.board.readBUSY()
}
