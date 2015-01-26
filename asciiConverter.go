package main

import (
//	"regexp"
)

func convertImageToHtml(image ImageFrame, width int) string {
	html := processCaca(image, width)
	return html
	//re := regexp.MustCompile(".*<body>(.*)</body>.*")
	//return re.ReplaceAllString(html, "$1")
}

func processCaca(img ImageFrame, width int) string {
	cols := width
	lines := int(cols * img.height * 6 / img.width / 10)

	canvas := NewCacaCanvas(0, 0)
	err := canvas.SetCanvasSize(cols, lines)
	if err != nil {
		fatal(err)
	}

	err = canvas.SetColorAnsi(CACA_BLACK, CACA_BLACK)
	if err != nil {
		fatal(err)
	}
	err = canvas.Clear()
	if err != nil {
		fatal(err)
	}

	dither := NewCacaDither(img.bpp, img.width, img.height)
	err = dither.SetAlgorithm("none")
	if err != nil {
		fatal(err)
	}
	err = dither.SetColor("fullgray")
	if err != nil {
		fatal(err)
	}

	err = dither.DitherImage(img, canvas)
	if err != nil {
		fatal(err)
	}

	output, err := canvas.ExportTo("html")
	if err != nil {
		fatal(err)
	}

	err = canvas.Free()
	if err != nil {
		fatal(err)
	}

	return output
}

// func processSimple(img ImageFrame) string {
// 	cols := int(120)
// 	lines := int(cols * img.height * 18 / img.width / 40)

// 	dw := int(img.width / cols)
// 	dh := int(img.height / lines)
// 	result := make([]byte, (cols+1)*lines)
// 	for y := 0; y < lines; y++ {
// 		for x := 0; x < cols; x++ {
// 			count := 0
// 			sx := x * dw
// 			sy := y * dh
// 			for dy := 0; dy < dw; dy++ {
// 				for dx := 0; dx < dh; dx++ {
// 					if sx+dx < img.width &&
// 						sy+dy < img.height &&
// 						img.Pixel(sx+dx, sy+dy) >= 128 {
// 						count++
// 					}
// 				}
// 			}
// 			if count >= dw*dh/2 {
// 				result[(cols+1)*y+x] = '$'
// 			} else {
// 				result[(cols+1)*y+x] = ' '
// 			}
// 		}
// 		result[(cols+1)*y+cols] = '\n'
// 	}
// 	return string(result)
// }
