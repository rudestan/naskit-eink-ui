package nasui

import (
	"errors"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"log"
	"nas-kit-ui/pkg/epd"
	"time"
)

const (
	DisplayTypePage = iota
	DisplayTypeMenu
)

const (
	DisplayWidth = 250
	DisplayHeight = 122
)

const (
	OrientationVertical = iota
	OrientationHorizontal
)

type NasUI struct {
	Epd *epd.Epaper
	Menu *Menu
	Pages []*Page
	BackgroundProc func(ctx *Context) error
	IndexPageName string
	Orientation int
	pageIndex int
	displayType int
	DefaultUI *DefaultUI
	currentPage *Page
	Debug bool
}

type Page struct {
	Name string
	RefreshInterval float64
	FullRedraw bool
	Display func(ctx *Context) (*image.RGBA, error)
	puCnt int
	drawnAt time.Time
}

type Menu struct {
	MenuItems []MenuItem
	Label string
	Page *Page
	ItemIndex int
	PerPage int
}

type MenuItem struct {
	Label string
	Page *Page
}

type Context struct {
	NasUI *NasUI
	DefaultUI *DefaultUI
}

var (
	ErrIndexPageNotFound = errors.New("index page not found")
	ErrNoPages = errors.New("no pages added to the ui")
)

func (p *Page) ResetCounters()  {
	p.drawnAt = time.Time{}
	p.puCnt = 0
}

func (p *Page) IsFirstTimeDisplay() bool {
	return p.puCnt == 0
}

func (p *Page) StopRefreshing()  {
	p.RefreshInterval = 0
}

func (p *Page) SetRefreshInterval(refreshInterval float64)  {
	p.RefreshInterval = refreshInterval
}

func (ui *NasUI) GetCurrentPage() *Page  {
	return ui.currentPage
}

func (ui *NasUI) getPageForButton(btn int) *Page  {
	switch btn {
	case epd.BtnOk:
		if ui.Menu != nil && ui.Menu.Page != nil {
			if ui.displayType == DisplayTypePage {
				ui.displayType = DisplayTypeMenu

				return ui.Menu.Page
			} else if ui.displayType == DisplayTypeMenu {
				menuItem := ui.Menu.MenuItems[ui.Menu.ItemIndex]

				if menuItem.Page != nil {
					ui.displayType = DisplayTypePage

					return menuItem.Page
				}
			}
		}
	case epd.BtnBack:
		ui.pageIndex = 0
		ui.Menu.ItemIndex = 0
		ui.displayType = DisplayTypePage

		return ui.Pages[0]
	case epd.BtnAdd: // previous
		if ui.displayType == DisplayTypePage {
			if ui.pageIndex-1 >= 0 {
				ui.pageIndex--
			} else {
				ui.pageIndex = len(ui.Pages) - 1
			}

			return ui.Pages[ui.pageIndex]
		} else if ui.displayType == DisplayTypeMenu {
			idx := ui.Menu.ItemIndex

			if idx - 1 >= 0 {
				idx--
			} else {
				idx = len(ui.Menu.MenuItems) - 1
			}

			ui.Menu.ItemIndex = idx

			return ui.Menu.Page
		}

	case epd.BtnSub: // next
		if ui.displayType == DisplayTypePage {
			if ui.pageIndex+1 < len(ui.Pages) {
				ui.pageIndex++
			} else {
				ui.pageIndex = 0
			}

			return ui.Pages[ui.pageIndex]
		} else if ui.displayType == DisplayTypeMenu {
			idx := ui.Menu.ItemIndex
			if idx + 1 < len(ui.Menu.MenuItems) {
				idx++
			} else {
				idx = 0
			}

			ui.Menu.ItemIndex = idx

			return ui.Menu.Page
		}
	}

	return nil
}

func (ui *NasUI) Run() error {
	if len(ui.Pages) == 0 {
		return ErrNoPages
	}

	idx := 0

	if ui.IndexPageName != "" {
		idx = ui.getIndexPage(ui.IndexPageName)
		if idx == -1 {
			return ErrIndexPageNotFound
		}
	}

	ui.pageIndex = idx
	activePage := ui.Pages[idx]

	if ui.Debug {
		img, _ := activePage.Display(ui.createContext())
		err := draw2dimg.SaveToPngFile("./debug.png", img)

		if err != nil {
			return err
		}

		return nil
	}

	err := ui.Epd.InitBoard()
	if err != nil {
		return err
	}

	err = ui.Epd.InitFull()
	if err != nil {
		return err
	}

	ui.Epd.Reset()
	ui.Epd.Clear(epd.BgColorWhite)

	ui.displayType = DisplayTypePage

	buttonsChan := make(chan int)
	errorChan := make(chan error)

	if ui.BackgroundProc != nil {
		go func() {
			err := ui.BackgroundProc(ui.createContext())

			if err != nil {
				errorChan <- err
			}
		}()
	}

	go func() {
		ui.Epd.ReadButtons(buttonsChan)
	}()

	go func() {
		for {
			select {
			case btn := <- buttonsChan:
				oldPage := activePage
				oldPage.ResetCounters()

				activePage = ui.getPageForButton(btn)
			default:
			}

			if activePage == nil {
				errorChan <- errors.New("no page do display")
			}

			if !activePage.drawnAt.IsZero() &&
				(activePage.RefreshInterval == 0 || time.Now().Sub(activePage.drawnAt).Seconds() < activePage.RefreshInterval) {
				continue
			}

			err := ui.displayPage(activePage)

			if err != nil {
				errorChan <- err
			}

			ui.currentPage = activePage
		}
	}()

	err = <- errorChan

	if err != nil {
		return err
	}

	return nil
}

func (ui *NasUI) AddPages(pages ...*Page)  {
	for _, page := range pages {
		ui.Pages = append(ui.Pages, page)
	}
}

func (ui *NasUI) createContext() *Context  {
	return &Context{
		NasUI: ui,
		DefaultUI:   ui.DefaultUI,
	}
}

func (ui *NasUI) displayPage(page *Page) error {
	defer func() {
		page.drawnAt = time.Now()
	}()

	img, err := page.Display(ui.createContext())

	if err != nil {
		return err
	}

	if img == nil {
		return nil
	}

	err = ui.initPage(page)
	if err != nil {
		return err
	}

	err = ui.Epd.Display(*img)
	if err != nil {
		return err
	}

	return nil
}

func (ui *NasUI) getIndexPage(name string) int  {
	for idx, page := range ui.Pages {
		if page.Name == name {
			return idx
		}
	}

	return -1
}

func (ui *NasUI) initPage(page *Page) error  {
	if page.puCnt > 1 {
		return nil
	}

	if page.puCnt == 0 {
		err := ui.Epd.InitFull()
		if err != nil {
			log.Println(err)
			return err
		}
		ui.Epd.Reset()
		ui.Epd.Clear(epd.BgColorWhite)

		page.puCnt++
	} else if page.puCnt == 1 && !page.FullRedraw {
		err := ui.Epd.InitPartial()
		if err != nil {
			log.Println(err)
			return err
		}
		page.puCnt++
	}

	return nil
}
