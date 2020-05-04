### Sunfounder NAS-Kit 2.13 eInk screen UI with drivers

Custom Go implementation of the UI for e-ink screen (which is actually HAT Waveshare 2.13 e-ink display) 
which is part of nice NAS-Kit by Sunfounder for building NAS with raspberryPI.

#### Why?

Sunfounder NAS-Kit goes with SD card which has preinstalled raspbean with some NAS Opensource software, also it includes
some python implementation of the UI: https://github.com/sunfounder/nas-kit. After some time playing around with the
software i was not really satisfied with its functionality and decided to just install minimal headless raspbean with SSH.
Of course this meant that cool e-ink screen will stop working, so it did. To restore the functionality it was enough to
just use the python source by Sunfounder. However, i did not like that in Sunfounder's implementation i cannot see
the stats for all of my attached HDD's and i have to install all the libs to the system that required by the implementation.
So i decided to implement the functionality that i need using Go also to be able to build one distributable binary.

#### Usage

To start the UI it is enough to just upload the binary to your raspberry PI with NAS-Kit and run it with needed arguments.
If you like the functionality and how UI looks and performs you can easily set this app binary to start automatically
once your rPI rebooted. For example add a start command line to `/etc/rc.local`.

##### Command line flags

| Flag          | Required| Description |
|---------------|---------|-------------|
| -d            | Yes     | Specify path to mounted disk(s) that you want to the stat for. To specify more than one mounting point - use multiple `-d` flags. You can list mounted disks for example with `df -aTh` command.|
| -ng           | No      | Do not group disk info by two on one page. If this flag specified every disk info will have it's own page.|
| -nf           | No      | Do not turn on the Fan if the temperature reaches 55°C|
| -p            | No      | Debug mode - will dump the current page to `debug.png` file. Can be used on local system to see how the UI image looks like.| 

#### Screenshots

|           | | 
|---------------|---------|
| <img src="https://github.com/rudestan/naskit-eink-ui/raw/master/screenshots/two_disks.png" width="250"> | Two disks info page. Will be shown if at least two disks specified with `-d` and no `-ng` flag.| 
| <img src="https://github.com/rudestan/naskit-eink-ui/raw/master/screenshots/single_disk.png" width="250"> | Single disk page is used in case  `-ng` flag specified or for odd numbers of disks.|
| <img src="https://github.com/rudestan/naskit-eink-ui/raw/master/screenshots/load.png" width="250"> | CPU load, temperature and RAM usage information.|
| <img src="https://github.com/rudestan/naskit-eink-ui/raw/master/screenshots/menu.png" width="250"> | Menu page (press "ok" button to open). From here you can reboot, power off rPI or check uptime info etc. On the top right corner you can see how many pages the menu has.|

#### Notes and Issues

- There are some issues and black artifacts if you change pages very often (would be good to fix)
- Fan is not using PWM however it is possible to code it in proper way using PWM
- Led Light is not used at all, and it is also possible to code it with PWM feature

#### Extra

This tiny project is for fun and might be never used by anyone at all :). However, it is easy to add a new page and display
any info on your NAS-Kit. I hope the code is self explainable because i do not have time to comment it. Feel free to drop me
a message if you have any questions. Clone and use the code as you want.



