package main

/*

#cgo pkg-config: caca
#include <stdlib.h>
#include <errno.h>
#include <caca.h>

*/
import "C"
import "unsafe"

var (
	CACA_BLACK       uint8 = C.CACA_BLACK
	CACA_DEFAULT     uint8 = C.CACA_DEFAULT
	CACA_TRANSPARENT uint8 = C.CACA_TRANSPARENT
)

func checkRet(ret C.int, err error) error {
	if int(ret) == 0 {
		return nil
	} else {
		return err
	}
}

type CacaCanvas struct {
	canvas *C.struct_caca_canvas
	cols   int
	lines  int
}

func NewCacaCanvas(cols int, lines int) *CacaCanvas {
	return &CacaCanvas{canvas: C.caca_create_canvas(C.int(cols), C.int(lines)), cols: cols, lines: lines}
}

func (this *CacaCanvas) Free() error {
	ret, err := C.caca_free_canvas(this.canvas)
	return checkRet(ret, err)
}

func (this *CacaCanvas) SetCanvasSize(cols int, lines int) error {
	ret, err := C.caca_set_canvas_size(this.canvas, C.int(cols), C.int(lines))
	this.cols = cols
	this.lines = lines
	return checkRet(ret, err)
}

func (this *CacaCanvas) SetColorAnsi(fg uint8, bg uint8) error {
	ret, err := C.caca_set_color_ansi(this.canvas, C.uint8_t(fg), C.uint8_t(bg))
	return checkRet(ret, err)
}

func (this *CacaCanvas) ExportTo(format string) (string, error) {
	ansi := C.CString(format)
	defer C.free(unsafe.Pointer(ansi))
	var length int
	ret, err := C.caca_export_canvas_to_memory(this.canvas, ansi, (*C.size_t)(unsafe.Pointer(&length)))
	if err != nil {
		return "", err
	}
	output := C.GoBytes(ret, C.int(length))
	return string(output), nil
}

func (this *CacaCanvas) Clear() error {
	ret, err := C.caca_clear_canvas(this.canvas)
	return checkRet(ret, err)
}

type CacaDither struct {
	dither *C.struct_caca_dither
}

func NewCacaDither(bpp int, width int, height int) *CacaDither {
	return &CacaDither{
		dither: C.caca_create_dither(C.int(bpp), C.int(width), C.int(height), C.int(bpp/8*width),
			C.uint32_t(0x00ff0000), C.uint32_t(0x0000ff00), C.uint32_t(0x000000ff), C.uint32_t(0x00000000))}
}

func (this *CacaDither) SetAlgorithm(algo string) error {
	algoCString := C.CString(algo)
	defer C.free(unsafe.Pointer(algoCString))
	ret, err := C.caca_set_dither_algorithm(this.dither, algoCString)
	return checkRet(ret, err)
}

func (this *CacaDither) SetColor(color string) error {
	colorCString := C.CString(color)
	defer C.free(unsafe.Pointer(colorCString))
	ret, err := C.caca_set_dither_color(this.dither, colorCString)
	return checkRet(ret, err)
}

func (this *CacaDither) DitherImage(image ImageFrame, canvas *CacaCanvas) error {
	ret, err := C.caca_dither_bitmap(canvas.canvas, 0, 0, C.int(canvas.cols), C.int(canvas.lines), this.dither, unsafe.Pointer(&image.data[0]))
	return checkRet(ret, err)
}