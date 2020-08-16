// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build dragonfly freebsd linux netbsd openbsd solaris
// +build !js
// +build !android

package devicescale

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type desktop int

const (
	desktopUnknown desktop = iota
	desktopGnome
	desktopCinnamon
	desktopUnity
	desktopKDE
	desktopXfce
)

func currentDesktop() desktop {
	tokens := strings.Split(os.Getenv("XDG_CURRENT_DESKTOP"), ":")
	switch tokens[len(tokens)-1] {
	case "GNOME":
		return desktopGnome
	case "X-Cinnamon":
		return desktopCinnamon
	case "Unity":
		return desktopUnity
	case "KDE":
		return desktopKDE
	case "XFCE":
		return desktopXfce
	default:
		return desktopUnknown
	}
}

var gsettingsRe = regexp.MustCompile(`\Auint32 (\d+)\s*\z`)

func gnomeScale() float64 {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "scaling-factor").Output()
	if err != nil {
		if err == exec.ErrNotFound {
			return 0
		}
		if _, ok := err.(*exec.ExitError); ok {
			return 0
		}
		panic(err)
	}
	m := gsettingsRe.FindStringSubmatch(string(out))
	s, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return float64(s)
}

func cinnamonScale() float64 {
	out, err := exec.Command("gsettings", "get", "org.cinnamon.desktop.interface", "scaling-factor").Output()
	if err != nil {
		if err == exec.ErrNotFound {
			return 0
		}
		if _, ok := err.(*exec.ExitError); ok {
			return 0
		}
		panic(err)
	}
	m := gsettingsRe.FindStringSubmatch(string(out))
	s, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return float64(s)
}

// custom bool for mapping "true"/"yes"/"on" to true
type XmlBool bool

// unmarshal from string to XmlBool
func (b *XmlBool) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	*b = XmlBool((s == "yes") || (s == "true") || (s == "on"))

	return nil
}

// cinnamon monitors config, strip out fields that are not needed
type cinnamonMonitorsConfig struct {
	XMLName       xml.Name `xml:"monitors"`
	Version       string   `xml:"version,attr"`
	Configuration struct {
		BaseScale float64 `xml:"base_scale"`
		Output    []struct {
			Scale    float64 `xml:"scale"`
			Primary  XmlBool `xml:"primary"`
		} `xml:"output"`
	} `xml:"configuration"`
}

const (
	//cinnamon default config file
	cinnamonCfg = ".config/cinnamon-monitors.xml"
)

func getCinnamonPrimaryMonitorScale() (scale float64) {
	scale = 0
	// getting current user is cached
	usr, _ := user.Current()
	if xmlFile, err := os.Open(filepath.Join(usr.HomeDir, cinnamonCfg)); err == nil {
		//goland:noinspection GoUnhandledErrorResult
		defer xmlFile.Close()
		if byteValue, err := ioutil.ReadAll(xmlFile); err == nil {
			var monitors cinnamonMonitorsConfig
			if err = xml.Unmarshal(byteValue, &monitors); err == nil {
				// base scale for all monitors
				scale = monitors.Configuration.BaseScale
				// for each monitor, get primary monitor scale
				for _, v := range monitors.Configuration.Output {
					if v.Primary {
						// we could not have a scale per monitor
						if v.Scale != 0.0 {
							scale = v.Scale
						}
						break
					}
				}
			}
		}
	}

	// retro-compatibility / fallback
	if scale == 0 {
		scale = cinnamonScale()
	}

	return
}

func impl(x, y int) float64 {
	s := -1.0
	switch currentDesktop() {
	case desktopGnome:
		// TODO: Support wayland and per-monitor scaling https://wiki.gnome.org/HowDoI/HiDpi
		s = gnomeScale()
	case desktopCinnamon:
		s = getCinnamonPrimaryMonitorScale()
	case desktopUnity:
		// TODO: Implement, supports per-monitor scaling
	case desktopKDE:
		// TODO: Implement, appears to support per-monitor scaling
	case desktopXfce:
		// TODO: Implement
	}
	if s <= 0 {
		s = 1
	}

	return s
}
