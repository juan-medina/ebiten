// Copyright 2020 The Ebiten Authors
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

// +build example jsgo

// Mascot is a desktop mascot on cross platforms.
// This is inspired by mattn's gopher (https://github.com/mattn/gopher).
package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"
	rmascot "github.com/hajimehoshi/ebiten/examples/resources/images/mascot"
)

const (
	width  = 200
	height = 200
)

var (
	gopher1 *ebiten.Image
	gopher2 *ebiten.Image
	gopher3 *ebiten.Image
)

func init() {
	// Decode image from a byte slice instead of a file so that
	// this example works in any working directory.
	// If you want to use a file, there are some options:
	// 1) Use os.Open and pass the file to the image decoder.
	//    This is a very regular way, but doesn't work on browsers.
	// 2) Use ebitenutil.OpenFile and pass the file to the image decoder.
	//    This works even on browsers.
	// 3) Use ebitenutil.NewImageFromFile to create an ebiten.Image directly from a file.
	//    This also works on browsers.
	img1, _, err := image.Decode(bytes.NewReader(rmascot.Out01_png))
	if err != nil {
		log.Fatal(err)
	}
	gopher1, _ = ebiten.NewImageFromImage(img1, ebiten.FilterDefault)

	img2, _, err := image.Decode(bytes.NewReader(rmascot.Out02_png))
	if err != nil {
		log.Fatal(err)
	}
	gopher2, _ = ebiten.NewImageFromImage(img2, ebiten.FilterDefault)

	img3, _, err := image.Decode(bytes.NewReader(rmascot.Out03_png))
	if err != nil {
		log.Fatal(err)
	}
	gopher3, _ = ebiten.NewImageFromImage(img3, ebiten.FilterDefault)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type dir int

const (
	right dir = iota
	left
)

type mascot struct {
	x16  int
	y16  int
	vy16 int

	dir   dir
	count int
}

func (m *mascot) update(screen *ebiten.Image) error {
	m.count++

	sw, sh := ebiten.ScreenSizeInFullscreen()
	ebiten.SetWindowPosition(m.x16/16, m.y16/16+sh-height)

	switch m.dir {
	case right:
		m.x16 += 64
	case left:
		m.x16 -= 64
	default:
		panic("not reached")
	}
	if m.x16/16 > sw-width && m.dir == right {
		m.dir = left
	}
	if m.x16 <= 0 && m.dir == left {
		m.dir = right
	}

	// Accelarate the mascot in the Y direction.
	m.vy16 += 8
	m.y16 += m.vy16

	// If the mascot is on the ground, stop it in the Y direction.
	if m.y16 >= 0 {
		m.y16 = 0
		m.vy16 = 0
	}

	// If the mascto is on the ground, cause an action in random.
	if rand.Intn(60) == 0 && m.y16 == 0 {
		switch rand.Intn(2) {
		case 0:
			// Jump.
			m.vy16 = -240
		case 1:
			// Turn.
			if m.dir == right {
				m.dir = left
			} else {
				m.dir = right
			}
		}
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	img := gopher1
	if m.y16 == 0 {
		switch (m.count / 3) % 4 {
		case 0:
			img = gopher1
		case 1, 3:
			img = gopher2
		case 2:
			img = gopher3
		}
	}
	op := &ebiten.DrawImageOptions{}
	if m.dir == left {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(width, 0)
	}
	screen.DrawImage(img, op)
	return nil
}

func main() {
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowFloating(true)
	m := &mascot{}
	if err := ebiten.Run(m.update, width, height, 1, "Mascot (Ebiten Demo)"); err != nil {
		log.Fatal(err)
	}
}
