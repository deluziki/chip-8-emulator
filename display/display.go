// Package display handles the graphical output for the CHIP-8 emulator using SDL2
package display

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	// CHIP-8 display dimensions
	Chip8Width  = 64
	Chip8Height = 32
)

// Display manages the SDL2 window and rendering
type Display struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	scale    int32
}

// New creates a new display with the specified scale factor
func New(title string, scale int32) (*Display, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		return nil, fmt.Errorf("failed to initialize SDL: %w", err)
	}

	window, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		Chip8Width*scale,
		Chip8Height*scale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		window.Destroy()
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	return &Display{
		window:   window,
		renderer: renderer,
		scale:    scale,
	}, nil
}

// Close cleans up SDL resources
func (d *Display) Close() {
	if d.renderer != nil {
		d.renderer.Destroy()
	}
	if d.window != nil {
		d.window.Destroy()
	}
	sdl.Quit()
}

// Clear clears the display with a black background
func (d *Display) Clear() {
	d.renderer.SetDrawColor(0, 0, 0, 255)
	d.renderer.Clear()
}

// Render draws the CHIP-8 display buffer to the screen
func (d *Display) Render(displayBuffer *[Chip8Width * Chip8Height]uint8) {
	d.Clear()

	// Set color for active pixels (white/green phosphor style)
	d.renderer.SetDrawColor(0, 255, 0, 255)

	for y := int32(0); y < Chip8Height; y++ {
		for x := int32(0); x < Chip8Width; x++ {
			if displayBuffer[y*Chip8Width+x] != 0 {
				rect := sdl.Rect{
					X: x * d.scale,
					Y: y * d.scale,
					W: d.scale,
					H: d.scale,
				}
				d.renderer.FillRect(&rect)
			}
		}
	}

	d.renderer.Present()
}

// SetTitle sets the window title
func (d *Display) SetTitle(title string) {
	d.window.SetTitle(title)
}
