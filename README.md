# oled-backlightd
Simple backlight daemon for oled laptops that system's backlight control don't work.

This little tool was created for controling the backlight brigthness on my Alienware 13R3. It watches the acpi backlight brigthness change and uses xrandr to adjust brightness.

## Usage:
go build
nohup ./oled-backlightd &

