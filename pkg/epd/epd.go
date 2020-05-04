package epd

import (
	"errors"
	"image"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	BtnOk = iota
	BtnBack
	BtnAdd
	BtnSub
)

// Epaper is a e-paper device
type Epaper struct {
	board  *board
	device device
	mu     sync.Mutex
}

// New creates a new e-paper device
func New(powerOffOnExit bool) *Epaper {
	b := &board{}
	d := newDev2in13(b)

	paper := &Epaper{
		board:  b,
		device: d,
	}

	if powerOffOnExit {
		paper.setExitSignalListener()
	}

	return paper
}

// Display display img on e-paper
func (p *Epaper) Display(img image.RGBA) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.board.isConnected() {
		return errors.New("the board is not connected, please call InitBoard first")
	}

	p.device.display(img)
	return nil
}

func (p *Epaper) InitBoard() error  {
	return p.device.initBoard()
}

func (p *Epaper) InitPartial() error  {
	return p.device.init(true)
}

func (p *Epaper) InitFull() error  {
	return p.device.init(false)
}

// Clear clear the e-paper screen
func (p *Epaper) Clear(bgColor byte) {
	p.device.clear(bgColor)
}

func (p *Epaper) Reset()  {
	p.device.reset()
}

func (p *Epaper) Sleep() {
	p.device.sleep()
}

func (p *Epaper) ReadButtons(btnChan chan int) {
	go func() {
		for {
			if p.board.keyOkPressed() {
				btnChan <- BtnOk
			}
		}
	}()

	go func() {
		for {
			if p.board.keyBackPressed() {
				btnChan <- BtnBack
			}
		}
	}()

	go func() {
		for {
			if p.board.keyAddPressed() {
				btnChan <- BtnAdd
			}
		}
	}()

	go func() {
		for {
			if p.board.keySubPressed() {
				btnChan <- BtnSub
			}
		}
	}()
}

func (p *Epaper) StartFan() {
	if p.board.isConnected() {
		p.board.writeFAN(true)
	}
}

func (p *Epaper) StopFan() {
	if p.board.isConnected() {
		p.board.writeFAN(false)
	}
}

func (p *Epaper) OnLed()  {
	p.board.writeLED(true)
}

func (p *Epaper) OffLed()  {
	p.board.writeLED(false)
}

func (p *Epaper) setExitSignalListener()  {
	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.board.cleanup()
		os.Exit(1)
	}()
}
