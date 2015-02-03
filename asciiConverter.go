package main

import "regexp"

type CacaContext struct {
	canvas *CacaCanvas
	dither *CacaDither
}

// cols is the number of chars in a row, ratio is the w/h of the canvas
func NewCacaContext(cols int, bpp int, width int, height int) (*CacaContext, error) {
	ctx := new(CacaContext)
	canvas := NewCacaCanvas(0, 0)
	dither := NewCacaDither(bpp, width, height)
	ctx.canvas = canvas
	ctx.dither = dither

	lines := int(cols * 5 / 10 * height / width)
	err := canvas.SetCanvasSize(cols, lines)
	if err != nil {
		ctx.Free()
		return nil, err
	}

	err = canvas.SetColorAnsi(CACA_WHITE, CACA_BLACK)
	if err != nil {
		ctx.Free()
		return nil, err
	}
	err = canvas.Clear()
	if err != nil {
		ctx.Free()
		return nil, err
	}

	err = dither.SetAlgorithm("none")
	if err != nil {
		ctx.Free()
		return nil, err
	}
	err = dither.SetColor("fullgray")
	if err != nil {
		ctx.Free()
		return nil, err
	}

	return ctx, nil
}

func (this *CacaContext) Free() {
	if this.canvas != nil {
		this.canvas.Free()
		this.canvas = nil
	}
	if this.dither != nil {
		this.dither.Free()
		this.dither = nil
	}
}

type AsciiConverter struct {
	cacaCtx *CacaContext
	re      *regexp.Regexp
}

func NewAsciiConverter(movie *Movie, cols int) (*AsciiConverter, error) {
	cacaCtx, err := NewCacaContext(cols, movie.Bpp, movie.Width, movie.Height)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(".*<body>(.*)</body>.*")
	if err != nil {
		return nil, err
	}
	return &AsciiConverter{cacaCtx, re}, nil
}

func (this *AsciiConverter) Free() {
	if this.cacaCtx != nil {
		this.cacaCtx.Free()
	}
}

func (this *AsciiConverter) ConvertToHtml(image *ImageFrame) (string, error) {
	html, err := processCaca(CACA_EXPORT_FMT_HTMLDIV, this.cacaCtx, image)
	if err != nil {
		return "", err
	}
	return html, nil
}

func (this *AsciiConverter) ConvertToAnsi(image *ImageFrame) (string, error) {
	text, err := processCaca(CACA_EXPORT_FMT_ANSI, this.cacaCtx, image)
	if err != nil {
		return "", err
	}
	return text, nil
}

func processCaca(format CacaExportFormat, ctx *CacaContext, img *ImageFrame) (string, error) {
	err := ctx.dither.DitherImage(img.Data, ctx.canvas)
	if err != nil {
		return "", err
	}

	output, err := ctx.canvas.ExportTo(format)
	if err != nil {
		return "", err
	}

	return output, nil
}

// Slow and simple implementation for testing and comparison
func processSimple(w int, h int, img *ImageFrame) string {
	cols := int(120)
	lines := int(cols * h * 10 / w / 17)

	dw := int(w / cols)
	dh := int(h / lines)
	result := make([]byte, (cols+1)*lines)
	data := img.Data
	for y := 0; y < lines; y++ {
		for x := 0; x < cols; x++ {
			count := 0
			sx := x * dw
			sy := y * dh
			for dy := 0; dy < dw; dy++ {
				for dx := 0; dx < dh; dx++ {
					if sx+dx < w && sy+dy < h && data[(sx+dx)*3+(sy+dy)*(w*3)] >= 128 {
						count++
					}
				}
			}
			if count >= dw*dh/2 {
				result[(cols+1)*y+x] = '$'
			} else {
				result[(cols+1)*y+x] = ' '
			}
		}
		result[(cols+1)*y+cols] = '\n'
	}
	return string(result)
}
