package epd

import (
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

const (
	// default gpio pins
	defaultPinRST  = "GPIO17"
	defaultPinDC   = "GPIO25"
	defaultPinCS   = "GPIO8"
	defaultPinBUSY = "GPIO24"
	defaultPinKeyOk = "GPIO5"
	defaultPinKeyBack = "GPIO6"
	defaultPinKeyAdd = "GPIO13"
	defaultPinKeySub = "GPIO19"
	defaultPinFan = "GPIO18"
	defaultPinLed = "GPIO26"
)

type board struct {
	connected bool
	port    spi.PortCloser
	spiConn spi.Conn
	// gpio pins
	pinRST  gpio.PinOut
	pinDC   gpio.PinOut
	pinCS   gpio.PinOut
	pinBUSY gpio.PinIn

	// button pins
	pinKeyOk   gpio.PinIn
	pinKeyBack gpio.PinIn
	pinKeyAdd    gpio.PinIn
	pinKeySub   gpio.PinIn

	// led and fan
	pinFan gpio.PinOut
	pinLed gpio.PinOut
}

func (p *board) init() error {
	_, err := host.Init()
	if err != nil {
		return err
	}

	p.port, err = spireg.Open("")
	if err != nil {
		return err
	}

	p.spiConn, err = p.port.Connect(2*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return err
	}

	p.pinRST = gpioreg.ByName(defaultPinRST)
	p.pinDC = gpioreg.ByName(defaultPinDC)
	p.pinCS = gpioreg.ByName(defaultPinCS)
	p.pinBUSY = gpioreg.ByName(defaultPinBUSY)

	p.pinKeyOk = gpioreg.ByName(defaultPinKeyOk)
	err = p.pinKeyOk.In(gpio.PullDown, gpio.BothEdges)
	if err != nil {
		return err
	}

	p.pinKeyBack = gpioreg.ByName(defaultPinKeyBack)
	err = p.pinKeyBack.In(gpio.PullDown, gpio.BothEdges)
	if err != nil {
		return err
	}

	p.pinKeyAdd = gpioreg.ByName(defaultPinKeyAdd)
	err = p.pinKeyAdd.In(gpio.PullDown, gpio.BothEdges)
	if err != nil {
		return err
	}

	p.pinKeySub = gpioreg.ByName(defaultPinKeySub)
	err = p.pinKeySub.In(gpio.PullDown, gpio.BothEdges)
	if err != nil {
		return err
	}

	p.pinLed = gpioreg.ByName(defaultPinLed)
	p.pinFan = gpioreg.ByName(defaultPinFan)

	p.connected = true

	return nil
}

func (p *board) isConnected() bool  {
	return p.connected
}

func (p *board) keyOkPressed() bool  {
	p.pinKeyOk.WaitForEdge(-1)

	return p.pinKeyOk.Read() == gpio.Low
}

func (p *board) keyBackPressed() bool  {
	p.pinKeyBack.WaitForEdge(-1)

	return p.pinKeyBack.Read() == gpio.Low
}

func (p *board) keyAddPressed() bool  {
	p.pinKeyAdd.WaitForEdge(-1)

	return p.pinKeyAdd.Read() == gpio.Low
}

func (p *board) keySubPressed() bool  {
	p.pinKeySub.WaitForEdge(-1)

	return p.pinKeySub.Read() == gpio.Low
}

func (p *board) ledOn()  {
	p.writeLED(true)
}

func (p *board) ledOff()  {
	p.writeLED(false)
}

func (p *board) writeLED(v bool)  {
	if v {
		p.pinLed.Out(gpio.High)
	} else {
		p.pinLed.Out(gpio.Low)
	}
}

func (p *board) fanOn()  {
	p.writeFAN(true)
}

func (p *board) fanOff()  {
	p.writeFAN(false)
}

func (p *board) writeFAN(v bool)  {
	if v {
		p.pinFan.Out(gpio.High)
	} else {
		p.pinFan.Out(gpio.Low)
	}
}

func (p *board) writeRST(v bool) {
	if v {
		p.pinRST.Out(gpio.High)
	} else {
		p.pinRST.Out(gpio.Low)
	}
}

func (p *board) writeDC(v bool) {
	if v {
		p.pinDC.Out(gpio.High)
	} else {
		p.pinDC.Out(gpio.Low)
	}
}

func (p *board) writeCS(v bool) {
	if v {
		p.pinCS.Out(gpio.High)
	} else {
		p.pinCS.Out(gpio.Low)
	}
}

func (p *board) readBUSY() {
	for p.pinBUSY.Read() == gpio.High {
		p.delayms(100)
	}
}

func (p *board) writeSPI(b ...byte) {
	r := make([]byte, len(b))
	p.spiConn.Tx(b, r)
}

func (p *board) cleanup() {
	defer func() {
		p.connected = false
	}()

	p.fanOff()
	p.ledOff()
	p.writeRST(false)
	p.writeDC(false)

	err := p.port.Close()

	if err != nil {
		log.Println(err)
	}
}

func (p *board) delayms(ms uint) {
	time.Sleep(time.Millisecond * time.Duration(ms))
}
