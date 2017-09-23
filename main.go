package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/rjeczalik/notify"
)

func main() {

	output := GetEmbeddedDP()
	if output == "" {
		log.Fatal("Failed to find any embeded DisplayPort device.")
	}
	backlight := GetACPIBacklight()
	if backlight == "" {
		log.Fatal("Failed to locate ACPI backlight directory.")
	}
	initbr := fmt.Sprintf("%.2f", GetCurrentBrightnessRatio(backlight))
	_, initerr := exec.Command("xrandr", "--output", output, "--brightness", initbr).Output()
	if initerr != nil {
		log.Fatal("Failed to set initial brightness.")
	}

	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(backlight+"/actual_brightness", c, notify.InModify); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	csignal := make(chan os.Signal, 1)
	signal.Notify(csignal, os.Interrupt, os.Kill)
	go func() {
		s := <-csignal
		notify.Stop(c)
		log.Fatal("Recieved signal: ", s)
	}()

	for {
		switch ei := <-c; ei.Event() {
		case notify.InModify:
			br := fmt.Sprintf("%.2f", GetCurrentBrightnessRatio(backlight))
			log.Println("Current brightness ", br)
			_, err := exec.Command("xrandr", "--output", output, "--brightness", br).Output()
			if err != nil {
				notify.Stop(c)
				log.Fatal("Failed to set brightness.")
			}
		}
	}
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
		log.Fatal("Failed to open max_brightness.")
	}
	_, err = fmt.Fscanf(fd, "%d", &maxBrightness)
	fd.Close()
	if err != nil {
		log.Fatal("Failed to read maximum brightness.")
	}

	var currentBrightness int
	fd, err = os.Open(backlight + "/actual_brightness")
	if err != nil {
		log.Fatal("Failed to open brightness.")
	}
	_, err = fmt.Fscanf(fd, "%d", &currentBrightness)
	fd.Close()
	if err != nil {
		log.Fatal("Failed to read brightness.")
	}
	ratio := float64(currentBrightness) / float64(maxBrightness)

	return ratio
}
