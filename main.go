package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	debug   = false
	screenX = 800
	screenY = 600

	boardX = 9
	boardY = 9
)

type game struct {
	ticks      int
	sampleJSON []byte
}

func (g *game) Update() error {
	g.ticks++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 64, 64, 255})
	s := fmt.Sprintf("Hello, wasmgame!\nTicks = %d\nThe content of asset/sample.json is: %s", g.ticks, string(g.sampleJSON))
	x, y := g.ticks%640, g.ticks%360
	ebitenutil.DebugPrintAt(screen, s, x, y)
}

func (g *game) Layout(w, h int) (int, int) {
	return 640, 360 // Screen resolution (not window size)
}

// open opens a file. In a browser, it downloads the file via HTTP;
// otherwise, it reads the file on disk.
func open(name string) (io.ReadCloser, error) {
	name = filepath.Clean(name)
	if runtime.GOOS == "js" {
		// TODO: use more lightweight method such as marwan-at-work/wasm-fetch
		resp, err := http.Get(name)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}

	return os.Open(name)
}

func readFile(name string) ([]byte, error) {
	f, err := open(name)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer f.Close()

	return io.ReadAll(f)
}

var ()

type SceneTransitionManager interface {
	SceneTransition(scene Scene)
}

type Scene interface {
	Update(manager SceneTransitionManager) error
	Draw(screen *ebiten.Image)
	init()
}

type Game struct {
	now_scene  Scene
	next_scene Scene
}

func (g *Game) Update() error {
	if g.next_scene != nil {
		g.now_scene = g.next_scene
		g.next_scene = nil
	}
	if err := g.now_scene.Update(g); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.now_scene.Draw(screen)
}

func (g *Game) SceneTransition(scene Scene) {
	g.next_scene = scene
	g.next_scene.init()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenX, screenY
}

func (g *Game) init() {
	imageSourceMap := map[string]string{
		"white_n":       "resources/images/go_white_n.png",
		"black_n":       "resources/images/go_black_n.png",
		"frame_white_n": "resources/images/go_frame_white_n.png",
		"frame_black_n": "resources/images/go_frame_black_n.png",
		"white_s":       "resources/images/go_white_s.png",
		"black_s":       "resources/images/go_black_s.png",
		"frame_white_s": "resources/images/go_frame_white_s.png",
		"frame_black_s": "resources/images/go_frame_black_s.png",
	}
	for key, value := range imageSourceMap {
		if err := loadImage(key, value); err != nil {
			log.Fatal(err)
		}

	}

	audioSourceMap := map[string]string{
		"set_stone":   "resources/se/set_stone.mp3",
		"force_stone": "resources/se/force_stone.mp3",
	}
	for key, value := range audioSourceMap {
		if err := loadAudio(key, value); err != nil {
			log.Fatal(err)
		}

	}
	playAudio("set_stone")
}

// NewGame method
func NewGame() *Game {
	g := &Game{}
	g.SceneTransition(&TitleScene{})
	g.init()
	return g
}

func main() {
	g := &game{}
	g.sampleJSON, _ = readFile("asset/sample.json")
	// ebiten.SetWindowSize(1280, 720) // has no effect on browser
	ebiten.SetWindowSize(screenX, screenY)
	ebiten.SetWindowTitle("Magnet Go!")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
