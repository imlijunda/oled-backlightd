package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {

	output := GetEmbeddedDP()
	if output == "" {
		fmt.Fprintf(os.Stderr, "Failed to find any embeded DisplayPort device.\n")
		os.Exit(1)
	}
	backlight := GetACPIBacklight()
	if backlight == "" {
		fmt.Fprintf(os.Stderr, "Failed to locate ACPI backlight directory.\n")
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					br := GetCurrentBrightnessRatio(backlight)
					sbr := fmt.Sprintf("%.2f", br)
					_, execerr := exec.Command("xrandr", "--output", output, "--brightness", sbr).Output()
					if execerr != nil {
						fmt.Fprintf(os.Stderr, "%s\n", execerr)
					}
				}
			case watchererr := <-watcher.Errors:
				fmt.Fprintf(os.Stderr, "%s", watchererr)
			}
		}
	}()
	err = watcher.Add(backlight + "/brightness")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	<-done
}

// GetEmbeddedDP Return the first embeded DisplayPort device reported by xrandr
func GetEmbeddedDP() string {

	cmd := exec.Command("xrandr")
	out, _ := cmd.Output()
	s := string(out)
	parsedOut := strings.Split(s, "\n")
	for i := 0; i <= len(parsedOut); i++ {
		tmp := strings.Split(parsedOut[i], " ")
		if (strings.HasPrefix(tmp[0], "eDP") || strings.HasPrefix(tmp[0], "e-DP")) && (tmp[1] == "connected") {
			return tmp[0]
		}
	}

	return ""
}

// GetACPIBacklight Return directory for ACPI backlight control
func GetACPIBacklight() string {

	d, _ := ioutil.ReadDir("/sys/class/backlight")
	for _, f := range d {
		fname := f.Name()
		if strings.HasPrefix(fname, "acpi") {
			return "/sys/class/backlight/" + fname
		}
	}

	return ""
}

// GetCurrentBrightnessRatio Return current blacklight level divided by max brightness
func GetCurrentBrightnessRatio(backlight string) float64 {

	var maxBrightness int
	fd, err := os.Open(backlight + "/max_brightness")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open max_brightness.\n")
	}
	_, err = fmt.Fscanf(fd, "%d", &maxBrightness)
	fd.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		fmt.Fprintf(os.Stderr, "Failed to read maximum brightness.\n")
		os.Exit(1)
	}

	var currentBrightness int
	fd, err = os.Open(backlight + "/brightness")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open brightness.\n")
	}
	_, err = fmt.Fscanf(fd, "%d", &currentBrightness)
	fd.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		fmt.Fprintf(os.Stderr, "Failed to read brightness.\n")
		os.Exit(1)
	}
	ratio := float64(currentBrightness) / float64(maxBrightness)

	return ratio
}
