package main

/*

#cgo pkg-config: caca
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <caca.h>

void *caca_export_html_div(caca_canvas_t const *cv, size_t *bytes)
{
    char *data, *cur;
    int x, y, len, w, h;

	w = caca_get_canvas_width(cv);
	h = caca_get_canvas_height(cv);

	*bytes = 1000 + h * (7 + w * (47 + 83 + 10 + 7));
    cur = data = malloc(*bytes);

    cur += sprintf(cur, "<div style=\"%s\">\n",
                        "font-family: monospace, fixed; font-weight: bold;");

    for(y = 0; y < h; y++)
    {
        const uint32_t *lineattr = caca_get_canvas_attrs(cv) + y * w;
        const uint32_t *linechar = caca_get_canvas_chars(cv) + y * w;

        for(x = 0; x < w; x += len)
        {
            cur += sprintf(cur, "<span style=\"");
            if(caca_attr_to_ansi_fg(lineattr[x]) != CACA_DEFAULT)
                cur += sprintf(cur, ";color:#%.03x",
                               caca_attr_to_rgb12_fg(lineattr[x]));
            if(caca_attr_to_ansi_bg(lineattr[x]) < 0x10)
                cur += sprintf(cur, ";background-color:#%.03x",
                               caca_attr_to_rgb12_bg(lineattr[x]));
            if(lineattr[x] & CACA_BOLD)
                cur += sprintf(cur, ";font-weight:bold");
            if(lineattr[x] & CACA_ITALICS)
                cur += sprintf(cur, ";font-style:italic");
            if(lineattr[x] & CACA_UNDERLINE)
                cur += sprintf(cur, ";text-decoration:underline");
            if(lineattr[x] & CACA_BLINK)
                cur += sprintf(cur, ";text-decoration:blink");
            cur += sprintf(cur, "\">");

            for(len = 0;
                x + len < w && lineattr[x + len] == lineattr[x];
                len++)
            {
                if(linechar[x + len] == CACA_MAGIC_FULLWIDTH)
                    ;
                else if((linechar[x + len] <= 0x00000020)
                        ||
                        ((linechar[x + len] >= 0x0000007f)
                         &&
                         (linechar[x + len] <= 0x000000a0)))
                {
					cur += sprintf(cur, "&#160;");
                }
                else if(linechar[x + len] == '&')
                    cur += sprintf(cur, "&amp;");
                else if(linechar[x + len] == '<')
                    cur += sprintf(cur, "&lt;");
                else if(linechar[x + len] == '>')
                    cur += sprintf(cur, "&gt;");
                else if(linechar[x + len] == '\"')
                    cur += sprintf(cur, "&quot;");
                else if(linechar[x + len] == '\'')
                    cur += sprintf(cur, "&#39;");
                else if(linechar[x + len] < 0x00000080)
                    cur += sprintf(cur, "%c", (uint8_t)linechar[x + len]);
                else if((linechar[x + len] <= 0x0010fffd)
                        &&
                        ((linechar[x + len] & 0x0000fffe) != 0x0000fffe)
                        &&
                        ((linechar[x + len] < 0x0000d800)
                         ||
                         (linechar[x + len] > 0x0000dfff)))
                    cur += sprintf(cur, "&#%i;", (unsigned int)linechar[x + len]);
                else
					cur += sprintf(cur, "&#%i;", (unsigned int)0x0000fffd);
            }
            cur += sprintf(cur, "</span>");
        }
        cur += sprintf(cur, "<br />\n");
    }

    cur += sprintf(cur, "</div></body></html>\n");

    *bytes = (uintptr_t)(cur - data);
    data = realloc(data, *bytes);

    return data;
}


*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	CACA_BLACK       uint8 = C.CACA_BLACK
	CACA_DEFAULT     uint8 = C.CACA_DEFAULT
	CACA_TRANSPARENT uint8 = C.CACA_TRANSPARENT
)

type CacaExportFormat uint8

const (
	CACA_EXPORT_FMT_ANSI CacaExportFormat = iota
	CACA_EXPORT_FMT_HTML
	CACA_EXPORT_FMT_HTMLDIV
)

var exportFmt []*C.char = []*C.char{
	C.CString("ansi"),
	C.CString("html"),
	C.CString("htmldiv"),
}

func (fmt CacaExportFormat) toCacaFmt() *C.char {
	return exportFmt[fmt]
}

type CacaCanvas struct {
	canvas *C.struct_caca_canvas
}

func NewCacaCanvas(cols int, lines int) *CacaCanvas {
	return &CacaCanvas{canvas: C.caca_create_canvas(C.int(cols), C.int(lines))}
}

func (this *CacaCanvas) Free() error {
	ret, err := C.caca_free_canvas(this.canvas)
	return checkRet(ret, err)
}

func (this *CacaCanvas) SetCanvasSize(cols int, lines int) error {
	ret, err := C.caca_set_canvas_size(this.canvas, C.int(cols), C.int(lines))
	return checkRet(ret, err)
}

func (this *CacaCanvas) Width() int {
	return int(C.caca_get_canvas_width(this.canvas))
}

func (this *CacaCanvas) Height() int {
	return int(C.caca_get_canvas_height(this.canvas))
}

func (this *CacaCanvas) SetColorAnsi(fg uint8, bg uint8) error {
	ret, err := C.caca_set_color_ansi(this.canvas, C.uint8_t(fg), C.uint8_t(bg))
	return checkRet(ret, err)
}

func (this *CacaCanvas) ExportTo(format CacaExportFormat) (string, error) {
	var length int
	var err error
	var ret unsafe.Pointer
	cacaFmt := format.toCacaFmt()
	switch format {
	case CACA_EXPORT_FMT_ANSI, CACA_EXPORT_FMT_HTML:
		ret, err = C.caca_export_canvas_to_memory(this.canvas, cacaFmt, (*C.size_t)(unsafe.Pointer(&length)))
	case CACA_EXPORT_FMT_HTMLDIV:
		ret, err = C.caca_export_html_div(this.canvas, (*C.size_t)(unsafe.Pointer(&length)))
	default:
		panic(fmt.Sprintf("Unsupported format: %d", int(format)))
	}
	if ret == nil || err != nil {
		if err == nil {
			return "", errors.New("Failed to export canvas to format: " + C.GoString(cacaFmt))
		}
		return "", err
	}
	defer C.free(ret)
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
		dither: C.caca_create_dither(C.int(bpp), C.int(height), C.int(height), C.int(bpp/8*width),
			C.uint32_t(0x00ff0000), C.uint32_t(0x0000ff00), C.uint32_t(0x000000ff), C.uint32_t(0x00000000))}
}

func (this *CacaDither) Free() error {
	ret, err := C.caca_free_dither(this.dither)
	return checkRet(ret, err)
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

func (this *CacaDither) DitherImage(data []byte, canvas *CacaCanvas) error {
	ret, err := C.caca_dither_bitmap(canvas.canvas,
		0, 0, C.int(canvas.Width()), C.int(canvas.Height()),
		this.dither, unsafe.Pointer(&data[0]))
	return checkRet(ret, err)
}
