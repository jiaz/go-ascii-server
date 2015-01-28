package main

import "github.com/jiaz/gmf"

// RGB packed image repr
type ImageFrame struct {
	Data []byte
}

type Movie struct {
	Width      int
	Height     int
	Bpp        int
	FrameCount int
	Images     <-chan *ImageFrame
}

func loadMovie(srcFileName string) (*Movie, error) {
	inputCtx, err := gmf.NewInputCtx(srcFileName)
	if err != nil {
		return nil, err
	}

	srcStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		return nil, err
	}

	srcCtx := srcStream.CodecCtx()
	w, h := srcCtx.Width(), srcCtx.Height()

	output := make(chan *ImageFrame)

	movie := new(Movie)
	movie.Width = w
	movie.Height = h
	movie.Bpp = 24
	movie.FrameCount = srcStream.NbFrames()
	movie.Images = output

	go func() {
		defer inputCtx.CloseInputAndRelease()
		defer close(output)

		dstCodec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_JPEG2000)
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
				swsCtx.Scale(frame, dstFrame)
				p := dstFrame.Data(0)
				output <- &ImageFrame{p}
			}
		}

	}()

	return movie, nil
}
