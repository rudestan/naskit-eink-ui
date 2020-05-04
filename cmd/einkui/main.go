package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"image"
	"log"
	"nas-kit-ui/pkg/epd"
	"nas-kit-ui/pkg/nasui"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type arrayFlags []string

func (af *arrayFlags) String() string {
	return "{}"
}

func (af *arrayFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}

func main()  {
	var diskFlags arrayFlags
	var notGroupFlag bool
	var noFanFlag bool
	var debugFlag bool

	flag.Var(&diskFlags, "d", "Partition label(s) to estimate the size")
	flag.BoolVar(&debugFlag, "p", false, "Debug UI and dump page to file")
	flag.BoolVar(&notGroupFlag, "ng", false, "Not group partitions")
	flag.BoolVar(&noFanFlag, "nf", false, "Do not use the FAN")
	flag.Parse()

	if len(diskFlags) == 0 {
		log.Fatal(errors.New("no partition(s) specified"))
	}

	fmt.Println("Creating UI")

	ui := createUi(debugFlag, noFanFlag)

	twoPathsCnt := len(diskFlags) / 2

	diskCounter := 0
	if !notGroupFlag && twoPathsCnt > 0 {
		diskFlags = addTwoPathsPages(diskFlags, &diskCounter, twoPathsCnt, ui)
	}

	if len(diskFlags) > 0 {
		addSinglePathPages(&diskCounter, diskFlags, ui)
	}

	addLoadPage(ui)

	err := ui.Run()

	if err != nil {
		log.Println(err)
	}
}

func gracefullShutdown(epaper *epd.Epaper)  {
	epaper.StopFan()

	err := epaper.InitFull()

	if err != nil {
		log.Fatal(err)
	}

	epaper.Clear(epd.BgColorWhite)
	epaper.Reset()
	epaper.Sleep()
}

func createUi(debugMode bool, noFan bool) *nasui.NasUI {
	ui := &nasui.NasUI{
		Debug: debugMode,
		DefaultUI: nasui.NewDefaultUI(nasui.OrientationVertical, "JetBrainsMono-Regular.ttf"),
		Epd: epd.New(true),
		Orientation: nasui.OrientationVertical,
		Menu:       &nasui.Menu{
			Label:     "Menu",
			Page: &nasui.Page{
				Name:            "Menu page",
				RefreshInterval: 2,
				Display: func(ctx *nasui.Context) (*image.RGBA, error) {
					return ctx.DefaultUI.MenuPage(ctx)
				},
			},
			MenuItems: []nasui.MenuItem{
				{
					Label: "Reboot Device",
					Page: &nasui.Page{
						RefreshInterval: 2,
						Display: func(ctx *nasui.Context) (*image.RGBA, error) {
							currentPage := ctx.NasUI.GetCurrentPage()

							if currentPage == nil {
								return nil, errors.New("no current page is set")
							}

							if !currentPage.IsFirstTimeDisplay() {
								gracefullShutdown(ctx.NasUI.Epd)

								cmd := exec.Command("bash", "-c", "sudo reboot")
								err := cmd.Run()

								return nil, err
							}

							return ctx.DefaultUI.MenuActionTextPage("Menu: reboot", []string{"Rebooting ... "}), nil
						},
					},
				},
				{
					Label: "Power off",
					Page: &nasui.Page{
						RefreshInterval: 2,
						Display: func(ctx *nasui.Context) (*image.RGBA, error) {
							currentPage := ctx.NasUI.GetCurrentPage()

							if currentPage == nil {
								return nil, errors.New("no current page is set!")
							}

							if !currentPage.IsFirstTimeDisplay() {
								gracefullShutdown(ctx.NasUI.Epd)

								cmd := exec.Command("bash", "-c", "sudo poweroff")
								err := cmd.Run()

								return nil, err
							}
							return ctx.DefaultUI.MenuActionTextPage("Menu: power off", []string{"Power off ... "}), nil
						},
					},
				},
				{
					Label: "Uptime",
					Page: &nasui.Page{
						RefreshInterval: 0.5,
						Display: func(ctx *nasui.Context) (*image.RGBA, error) {
							cmd := exec.Command("bash", "-c", "uptime -p")
							out, err := cmd.Output()

							if err != nil {
								return nil, err
							}

							uptimeStrs := strings.Split(string(out), ",")

							return ctx.DefaultUI.MenuActionTextPage("Menu: uptime", uptimeStrs), nil
						},
					},
				},
				{
					Label: "Exit",
					Page: &nasui.Page{
						RefreshInterval: 0.5,
						Display: func(ctx *nasui.Context) (*image.RGBA, error) {
							currentPage := ctx.NasUI.GetCurrentPage()

							if currentPage == nil {
								return nil, errors.New("no current page set")
							}

							if !currentPage.IsFirstTimeDisplay() {
								gracefullShutdown(ctx.NasUI.Epd)

								os.Exit(0)

								return nil, nil
							}

							return ctx.DefaultUI.MenuActionTextPage("Menu: exit", []string{"Bye!"}), nil
						},
					},
				},
			},
		},
	}

	if !noFan {
		ui.BackgroundProc = func(ctx *nasui.Context) error {
			for {
				temp, err := getCpuTemp()

				if err != nil {
					return err
				}

				if temp >= 55 {
					ctx.NasUI.Epd.StartFan()
				}

				if temp <= 43 {
					ctx.NasUI.Epd.StopFan()
				}

				time.Sleep(2 * time.Second)
			}
		}
	}

	return ui
}

func addTwoPathsPages(diskFlags arrayFlags, discCounter *int, twoPathsCnt int, ui *nasui.NasUI) arrayFlags  {
	for i := 0; i < twoPathsCnt; i++ {
		diskPaths := []string{diskFlags[i], diskFlags[i+1]}

		label := fmt.Sprintf("Disk %d&%d", *discCounter+1, *discCounter+2)
		*discCounter += 2

		ui.AddPages(&nasui.Page{
			Name:            label,
			RefreshInterval: 0.8,
			Display: func(ctx *nasui.Context) (*image.RGBA, error) {
				partitionStat, err := getPartitionStat(diskPaths)

				if err != nil {
					return nil, err
				}

				if len(partitionStat) < 2 {
					log.Fatal(fmt.Sprintf("Partitions \"%s\" not found or stat not avaliable", diskPaths))
				}

				diskStats := []*nasui.DiskInfo{
					{
						Idx: 		 fmt.Sprintf("%d", i),
						Path:        partitionStat[0].Path,
						Total:       humanize.Bytes(partitionStat[0].Total),
						Free:        humanize.Bytes(partitionStat[0].Free),
						Used:        humanize.Bytes(partitionStat[0].Used),
						UsedPercent: partitionStat[0].UsedPercent,
					},
					{
						Idx: 		 fmt.Sprintf("%d", i+1),
						Path:        partitionStat[1].Path,
						Total:       humanize.Bytes(partitionStat[1].Total),
						Free:        humanize.Bytes(partitionStat[1].Free),
						Used:        humanize.Bytes(partitionStat[1].Used),
						UsedPercent: partitionStat[1].UsedPercent,
					},
				}

				img, err := ctx.DefaultUI.DiscInfoTwoDiscs(label, getOutboundIP().String(), diskStats)

				if err != nil {
					return nil, err
				}

				return img, nil
			},
		})

		diskFlags = remove(diskFlags, i+1)
		diskFlags = remove(diskFlags, i)
	}

	return diskFlags
}

func addSinglePathPages(diskCounter *int, diskFlags arrayFlags, ui *nasui.NasUI)  {
	for _, diskFlag := range diskFlags {
		*diskCounter++
		label := fmt.Sprintf("Disk %d", *diskCounter)
		ui.AddPages(&nasui.Page{
			Name:            label,
			RefreshInterval: 0.8,
			Display: func(ctx *nasui.Context) (*image.RGBA, error) {
				partitionStat, err := getPartitionStat([]string{diskFlag})

				if err != nil {
					return nil, err
				}

				if len(partitionStat) != 1 {
					log.Fatal(fmt.Sprintf("Partition \"%s\" not found or stat not avaliable", diskFlag))
				}

				img, err := ctx.DefaultUI.DiscInfoOneDisc(
					label,
					getOutboundIP().String(),
					&nasui.DiskInfo{
						Path:        partitionStat[0].Path,
						Total:       humanize.Bytes(partitionStat[0].Total),
						Free:        humanize.Bytes(partitionStat[0].Free),
						Used:        humanize.Bytes(partitionStat[0].Used),
						UsedPercent: partitionStat[0].UsedPercent,
					})

				if err != nil {
					return nil, err
				}

				return img, nil
			},
		})
	}
}

func addLoadPage(ui *nasui.NasUI) {
	ui.AddPages(&nasui.Page{
		Name:            "Load",
		RefreshInterval: 0.5,
		Display: func(ctx *nasui.Context) (*image.RGBA, error) {
			cpuPercent, err := cpu.Percent(0, false)
			if err != nil {
				return nil, err
			}

			memInfo, err := mem.VirtualMemory()
			if err != nil {
				return nil, err
			}

			temp, err := getCpuTemp()

			if err != nil {
				return nil, err
			}

			usageInfo := &nasui.UsageInfo{
				CpuPercent: fmt.Sprintf("%.2f%%", cpuPercent[0]),
				CpuTemp:    fmt.Sprintf("%.2f°C", temp),
				RamPercent: fmt.Sprintf("%.2f%%", memInfo.UsedPercent),
				RamUsed:    humanize.Bytes(memInfo.Used),
			}

			return ctx.DefaultUI.ResourcesInfo("Usage", getOutboundIP().String(), usageInfo)
		},
	})
}

func remove(diskFlags arrayFlags, s int) arrayFlags {
	return append(diskFlags[:s], diskFlags[s+1:]...)
}

func getPartitionStat(paths []string) ([]*disk.UsageStat, error) {
	var usageStats []*disk.UsageStat

	dp, err := disk.Partitions(false)

	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if mountPointExists(path, dp) {
			usageStat, err := disk.Usage(path)

			if err != nil {
				continue
			}

			usageStats = append(usageStats, usageStat)
		}
	}

	return usageStats, nil
}

func getCpuTemp() (float64, error) {
	cmd := exec.Command("bash", "-c", "cat /sys/class/thermal/thermal_zone0/temp")
	out, err := cmd.Output()

	if err != nil {
		return 0, err
	}

	realTemp, err := strconv.ParseFloat(strings.Trim(string(out), "\n"), 64)

	if err != nil {
		return 0, err
	}

	return realTemp / 1000, nil
}

func mountPointExists(path string, stats []disk.PartitionStat) bool  {
	for _, stat := range stats {
		if stat.Mountpoint == path {
			return true
		}
	}

	return false
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = conn.Close()

		if err != nil {
			log.Fatal(errors.New("failed to get an IP"))
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
