package main

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/jiaz/gmf"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
	os.Exit(0)
}

// RGB packed image repr
type ImageFrame struct {
	data   []byte
	width  int
	height int
	bpp    int
}

func loadingMov(srcFileName string) <-chan ImageFrame {
	output := make(chan ImageFrame)

	go func() {
		inputCtx, _ := gmf.NewInputCtx(srcFileName)
		defer inputCtx.CloseInputAndRelease()

		srcStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
		if err != nil {
			fatal(err)
		}

		srcCtx := srcStream.CodecCtx()
		w, h := srcCtx.Width(), srcCtx.Height()

		dstCodec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_PNG)
		if err != nil {
			fatal(err)
		}
		dstCtx := gmf.NewCodecCtx(dstCodec)
		defer gmf.Release(dstCtx)

		dstCtx.SetPixFmt(gmf.AV_PIX_FMT_RGB24).SetWidth(w).SetHeight(h)
		if dstCodec.IsExperimental() {
			dstCtx.SetStrictCompliance(-2)
		}
		if err := dstCtx.Open(nil); err != nil {
			fatal(err)
		}

		swsCtx := gmf.NewSwsCtx(srcCtx, dstCtx, gmf.SWS_POINT)
		defer gmf.Release(swsCtx)

		dstFrame := gmf.NewFrame().SetWidth(w).SetHeight(h).SetFormat(gmf.AV_PIX_FMT_RGB24)
		defer gmf.Release(dstFrame)

		if err := dstFrame.ImgAlloc(); err != nil {
			fatal(err)
		}

		for packet := range inputCtx.GetNewPackets() {
			defer gmf.Release(packet)

			if packet.StreamIndex() != srcStream.Index() {
				continue
			}

			inStream, err := inputCtx.GetStream(packet.StreamIndex())
			if err != nil {
				fatal(err)
			}

			inCtx := inStream.CodecCtx()

			for frame := range packet.Frames(inCtx) {
				defer gmf.Release(frame)

				swsCtx.Scale(frame, dstFrame)
				p := dstFrame.Data(0)
				output <- ImageFrame{p, w, h, 24}
			}
		}

		close(output)

	}()

	return output
}
