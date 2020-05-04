package nasui

import (
	"bytes"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/png"
	"log"
	"math"
	"strings"
)

type DefaultUI struct {
	width int
	height int
	font string
}

type DiskInfo struct {
	Idx string
	Path string
	Total string
	Free string
	Used string
	UsedPercent float64
}

type UsageInfo struct {
	CpuPercent string
	CpuTemp string
	RamPercent string
	RamUsed string
}

const maxMenuItemsPerPage = 3

func NewDefaultUI(orientation int, font string) *DefaultUI  {
	width := DisplayWidth
	height := DisplayHeight

	if orientation == OrientationHorizontal {
		width = DisplayHeight
		height = DisplayWidth
	}

	defaultUi := &DefaultUI{
		width:  width,
		height: height,
		font: font,
	}

	err := defaultUi.registerFont()

	if err != nil {
		log.Fatal(err)
	}

	return defaultUi
}

func (de *DefaultUI) MenuActionTextPage(label string, lines []string) *image.RGBA  {
	dest := image.NewRGBA(image.Rect(0, 0, de.width, de.height)) // horizontal
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(image.White)
	drawRect(gc, 0, 0, float64(de.width), float64(de.height))
	gc.Fill()

	gc.SetFillColor(image.Black)
	gc.SetFontSize(14)
	gc.SetFontData(draw2d.FontData{
		Name: de.font,
	})

	row := 16.0
	gc.FillStringAt(label, 0, row)

	gc.SetFillColor(image.Black)
	drawRect(gc, 0, 18, float64(de.width), 1)
	gc.Fill()

	gc.SetFontSize(14)

	base := 40.0

	for idx, line := range lines {
		space := float64(idx) * row + float64(idx) * 4

		gc.FillStringAt(strings.Trim(line, "\n "), 14, base + space)
	}

	return dest
}

func (de *DefaultUI) MenuPage(ctx *Context) (*image.RGBA, error) {
	dest := image.NewRGBA(image.Rect(0, 0, de.width, de.height)) // horizontal
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(image.White)
	drawRect(gc, 0, 0, float64(de.width), float64(de.height))
	gc.Fill()

	gc.SetFillColor(image.Black)
	gc.SetFontSize(14)

	gc.SetFontData(draw2d.FontData{
		Name: de.font,
	})

	itemsPerPage := ctx.NasUI.Menu.PerPage

	if itemsPerPage > maxMenuItemsPerPage || itemsPerPage <= 0 {
		itemsPerPage = maxMenuItemsPerPage
	}

	row := 16.0
	totalPages := int(math.Ceil(float64(len(ctx.NasUI.Menu.MenuItems)) / float64(itemsPerPage)))
	pageN := 1
	offset := 0
	space := 2

	if len(ctx.NasUI.Menu.MenuItems) > itemsPerPage && ctx.NasUI.Menu.ItemIndex > itemsPerPage - 1 {
		pageN = int(math.Ceil(float64(ctx.NasUI.Menu.ItemIndex + 1) / float64(itemsPerPage)))
		offset = itemsPerPage * (pageN - 1)
	}

	gc.FillStringAt(ctx.NasUI.Menu.Label, 0, row)
	gc.FillStringAt(fmt.Sprintf("%d/%d", pageN, totalPages), float64(de.width - 44), row)


	gc.SetFillColor(image.Black)
	drawRect(gc, 0, 18, float64(de.width), 1)
	gc.Fill()

	for n := 1; n <= itemsPerPage; n++ {

		idx := n - 1 + offset
		if idx > len(ctx.NasUI.Menu.MenuItems) - 1 {
			break
		}

		menuItem := ctx.NasUI.Menu.MenuItems[idx]

		if idx == ctx.NasUI.Menu.ItemIndex {
			gc.SetFillColor(image.Black)
			drawRect(gc, 8, float64(n * 28 + space * n), float64(de.width-16), 28)
			gc.Fill()

			gc.SetFillColor(image.White)
			drawRect(gc, 9, float64(n * 30 + space - 1), float64(de.width-18), 26)
			gc.Fill()

			gc.SetFillColor(image.Black)
		}

		gc.SetFontSize(14)
		gc.FillStringAt(fmt.Sprintf("%d.%s", idx + 1, menuItem.Label), 14, float64(n * 30 + 19))
	}

	return dest, nil
}

func (de *DefaultUI) DiscInfoOneDisc(label string, bgLabel string, di *DiskInfo) (*image.RGBA, error)  {
	dest := image.NewRGBA(image.Rect(0, 0, de.width, de.height))
	gc := de.newClearGC(dest)

	de.AddPageHeader(gc, label, bgLabel)

	row := 14.0
	space := 2.0

	gc.SetFillColor(image.Black)
	gc.SetFontSize(14)
	gc.FillStringAt(fmt.Sprintf("Path: %s", di.Path), 0, (row + space) * 2 + 12)

	// Gauge:
	gc.SetFillColor(image.Black)
	drawRect(gc, 0, 54, float64(de.width), 18)
	gc.Fill()

	gc.SetFillColor(image.White)
	drawRect(gc, 2, 56, float64(de.width - 4), 14)
	gc.Fill()


	gaugeLength := float64(de.width - 3) / 100 * di.UsedPercent

	if gaugeLength > float64(de.width - 3) {
		gaugeLength = float64(de.width - 3)
	}

	if gaugeLength < 3 {
		gaugeLength = 3
	}

	gc.SetFillColor(image.Black)
	drawRect(gc, 3, 57, gaugeLength - 3, 12)
	gc.Fill()
	// eo gauge

	gc.SetFillColor(image.Black)
	gc.FillStringAt(fmt.Sprintf("U: %s from %s", di.Used, di.Total), 0, (row + space) * 5 + 10)

	// Total
	gc.SetFillColor(image.Black)
	drawRect(gc, 0, float64(de.height - 20), 110, 20)
	gc.Fill()

	gc.SetFillColor(image.White)
	gc.FillStringAt(fmt.Sprintf("F: %s", di.Free), 2, float64(de.height - 5))

	return dest, nil
}

func (de *DefaultUI) DiscInfoTwoDiscs(label string, bgLabel string, dis []*DiskInfo) (*image.RGBA, error)  {
	dest := image.NewRGBA(image.Rect(0, 0, de.width, de.height))
	gc := de.newClearGC(dest)

	de.AddPageHeader(gc, label, bgLabel)

	infoPanHeight := 52
	row := 14.0
	space := 2.0

	gc.SetFontSize(12)

	for idx, di := range dis {
		panHeight := float64(infoPanHeight * idx)

		gc.SetFillColor(image.Black)
		drawRect(gc, 0, (row + space) + 6 + panHeight, 16, 16)
		gc.Fill()

		gc.SetFillColor(image.White)
		gc.FillStringAt(di.Idx, 3, (row + space) * 2 + 4 + panHeight)

		gc.SetFillColor(image.Black)
		gc.FillStringAt(di.Path, 18, (row + space) * 2 + 4 + panHeight)

		// Gauge:
		gc.SetFillColor(image.Black)
		drawRect(gc, 0, 39 + panHeight, float64(de.width), 14)
		gc.Fill()

		gc.SetFillColor(image.White)
		drawRect(gc, 2, 41 + panHeight, float64(de.width - 4), 10)
		gc.Fill()

		gaugeLength := float64(de.width - 3) / 100 * di.UsedPercent

		if gaugeLength > float64(de.width - 3) {
			gaugeLength = float64(de.width - 3)
		}

		if gaugeLength < 3 {
			gaugeLength = 3
		}

		gc.SetFillColor(image.Black)
		drawRect(gc, 3, 42 + panHeight, gaugeLength - 3, 8)
		gc.Fill()
		// eo gauge

		gc.SetFillColor(image.Black)
		gc.FillStringAt(
			fmt.Sprintf(
				"%s/%s F:%s",
				strings.ReplaceAll(di.Used, " ", ""),
				strings.ReplaceAll(di.Total, " ", ""),
				strings.ReplaceAll(di.Free, " ", "")), 0, (row + space) * 4 + 3 + panHeight)
	}

	return dest, nil
}

func (de *DefaultUI) ResourcesInfo(label string, bgLabel string, usageInfo *UsageInfo) (*image.RGBA, error) {
	dest := image.NewRGBA(image.Rect(0, 0, de.width, de.height))
	gc := de.newClearGC(dest)

	de.AddPageHeader(gc, label, bgLabel)

	// cpu

	br := bytes.NewReader(IconCpu)
	gcIcon, err := png.Decode(br)

	if err != nil {
		return nil, err
	}

	gc.SetFontSize(14)

	gc.Translate(10,32)
	gc.DrawImage(gcIcon)

	gc.SetFillColor(image.Black)


	gc.FillStringAt(fmt.Sprintf("%s, %s", usageInfo.CpuPercent, usageInfo.CpuTemp), 44, 22)

	// ram
	br = bytes.NewReader(IconRam)
	gcIcon, err = png.Decode(br)
	if err != nil {
		return nil, err
	}

	gc.Translate(0,42)
	gc.DrawImage(gcIcon)
	gc.SetFillColor(image.Black)
	gc.FillStringAt(fmt.Sprintf("%s, %s", usageInfo.RamPercent, usageInfo.RamUsed), 44, 22)

	return dest, nil
}

func (de *DefaultUI) newClearGC(img *image.RGBA) *draw2dimg.GraphicContext  {
	gc := draw2dimg.NewGraphicContext(img)

	gc.SetFillColor(image.White)
	drawRect(gc, 0, 0, float64(de.width), float64(de.height))
	gc.Fill()

	return gc
}

func (de *DefaultUI) AddPageHeader(gc *draw2dimg.GraphicContext, label string, bgLabel string)  {
	gc.SetFillColor(image.Black)
	gc.SetFontSize(14)

	gc.SetFontData(draw2d.FontData{
		Name: de.font,
	})

	row := 16.0

	gc.FillStringAt(label, 0, row)

	gc.SetFillColor(image.Black)
	drawRect(gc, 110, 0, float64(de.width - 110), row + 4)
	gc.Fill()

	gc.SetFillColor(image.White)
	gc.FillStringAt(bgLabel, 116, row)

	gc.SetFillColor(image.Black)
	drawRect(gc, 0, row + 4, float64(de.width), 1)
	gc.Fill()
}

func drawRect(gc *draw2dimg.GraphicContext, x, y, w, h float64) {
	gc.BeginPath()
	gc.MoveTo(x, y)
	gc.LineTo(x+w, y)
	gc.LineTo(x+w, y+h)
	gc.LineTo(x, y+h)
	gc.Close()
}

func (de *DefaultUI) registerFont() error {
	font, err := truetype.Parse(DefaultFont)

	if err != nil {
		return err
	}

	draw2d.RegisterFont(draw2d.FontData{
		Name: de.font,
	}, font)

	return nil
}
