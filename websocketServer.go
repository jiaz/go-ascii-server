package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"golang.org/x/net/websocket"
)

type CachingData struct {
	VideoBuffer []string
	FrameCount  int
}

var (
	indexTmpl *template.Template
	data      *CachingData
)

func readFromCache(cacheFilePath string) {
	cacheFile, err := os.Open(cacheFilePath)
	if err != nil {
		fatal(err)
		// TODO: can fall back to read from origin and try recreate the cache
	}

	dec := gob.NewDecoder(cacheFile)
	localData := CachingData{}
	err = dec.Decode(&localData)
	if err != nil {
		fatal(err)
	}
	data = &localData
}

func writeToCache(cacheFilePath string) {
	cacheFileTmp := cacheFilePath + ".tmp"
	cacheFileTmpFile, err := os.Create(cacheFileTmp)
	if err != nil {
		fatal(err)
	}
	enc := gob.NewEncoder(cacheFileTmpFile)
	err = enc.Encode(data)
	if err != nil {
		fatal(err)
	}
	if err = os.Rename(cacheFileTmp, cacheFilePath); err != nil {
		fatal(err)
	}
}

func warmUp() {
	log.Println("warming up server...")
	moviePath := filepath.Join(config.ResourcesPath, "demo.m4v")
	cachePath := moviePath + ".cache"

	if ok, err := LockFile(cachePath); !ok {
		fatal(err)
	}
	defer func() {
		if ok, err := UnlockFile(cachePath); !ok {
			fatal(err)
		}
	}()

	if ok, err := checkFileExists(cachePath); ok {
		readFromCache(cachePath)
	} else if err == nil {
		movie, err := loadMovie(moviePath)
		if err != nil {
			fatal(err)
		}
		converter, err := NewAsciiConverter(movie, 120)
		if err != nil {
			fatal(err)
		}
		defer converter.Free()

		data = new(CachingData)
		data.VideoBuffer = make([]string, movie.FrameCount)
		data.FrameCount = 0
		for image := range movie.ImageStream {
			htmlDiv, err := converter.ConvertToHtml(image)
			if err != nil {
				fatal(err)
			}

			var b bytes.Buffer
			w := gzip.NewWriter(&b)
			if _, err := w.Write([]byte(htmlDiv)); err != nil {
				fatal(err)
			}
			w.Close()

			data.VideoBuffer[data.FrameCount] = b.String()
			// data.VideoBuffer[data.FrameCount] = htmlDiv

			if data.FrameCount%100 == 0 {
				log.Println("Loading frame:", data.FrameCount)
			}
			data.FrameCount++
		}
		writeToCache(cachePath)
	} else {
		fatal(err)
	}
	log.Println("warming up done")
}

func root(w http.ResponseWriter, r *http.Request) {
	var model = struct{ Host string }{config.WebsocketHost}
	err := indexTmpl.Execute(w, model)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

type WSRequest struct {
	Type string
	Args map[string]interface{}
}

type WSResponse struct {
	ErrorCode int
	Type      string
	Data      map[string]interface{}
}

type SendDataArgs struct {
	FromFrame int
	ToFrame   int
}

func (this *SendDataArgs) Load(cmd *WSRequest) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	this.FromFrame = int(cmd.Args["from"].(float64))
	this.ToFrame = int(cmd.Args["to"].(float64))
	return nil
}

func sendData(conn *websocket.Conn, args *SendDataArgs) {
	log.Println("Start streaming, from:", args.FromFrame, "to:", args.ToFrame)

	if args.FromFrame < 0 || args.FromFrame >= args.ToFrame || args.ToFrame > data.FrameCount {
		log.Println("Illegal frame numbers")
		sendError(conn, "GETDATA", errors.New("Invalid range"))
		return
	}

	for _, frame := range data.VideoBuffer[args.FromFrame:args.ToFrame] {
		str := base64.StdEncoding.EncodeToString([]byte(frame))
		websocket.JSON.Send(conn, WSResponse{200, "GETDATA", map[string]interface{}{"Frame": str}})
		// websocket.JSON.Send(conn, WSResponse{200, "GETDATA", map[string]interface{}{"Frame": frame}})
	}

	log.Println("Finished streaming, from:", args.FromFrame, "to:", args.ToFrame)
}

func sendFrameCount(conn *websocket.Conn) {
	log.Println("Send frame count:", data.FrameCount)
	websocket.JSON.Send(conn, WSResponse{200, "GETFRAMECOUNT", map[string]interface{}{"FrameCount": data.FrameCount}})
	log.Println("Finish send frame count")
}

func sendError(conn *websocket.Conn, cmdType string, err error) {
	log.Println("Send error:", err)
	websocket.JSON.Send(conn, WSResponse{500, cmdType, map[string]interface{}{"Err": err.Error()}})
	log.Println("Finish send error")
}

func workingProc(conn *websocket.Conn, cmdQueue <-chan *WSRequest, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		if cmd, more := <-cmdQueue; !more {
			break
		} else {
			// process cmd
			switch cmd.Type {
			case "GETDATA":
				args := new(SendDataArgs)
				if err := args.Load(cmd); err != nil {
					sendError(conn, cmd.Type, err)
				} else {
					sendData(conn, args)
				}
			case "GETFRAMECOUNT":
				sendFrameCount(conn)
			default:
				sendError(conn, cmd.Type, errors.New(fmt.Sprintf("Unknown command: %#v", cmd)))
			}
		}
	}
}

func websocketHandler(conn *websocket.Conn) {
	log.Println("Begin serving connection:", conn)
	var wg sync.WaitGroup
	commandQueue := make(chan *WSRequest, 10)
	go workingProc(conn, commandQueue, &wg)
	for {
		// try read command from conn
		var cmd WSRequest
		if err := websocket.JSON.Receive(conn, &cmd); err != nil {
			if err == io.EOF {
				log.Println("EOF")
				break
			} else {
				log.Println("Error receving command:", err, ". Close connection")
				break
			}
		} else {
			log.Println("Get cmd", cmd)
			commandQueue <- &cmd
		}
	}
	log.Println("Closing command queue")
	close(commandQueue)
	log.Println("Queue closed. Waiting for clean up")
	// wait for clean up
	wg.Wait()
	log.Println("Finish serving connection:", conn)
}

func NewPlayerServer() *websocket.Server {
	return &websocket.Server{Handler: websocketHandler}
}

func serveStatic(folder string) {
	http.Handle(
		"/"+folder+"/",
		http.StripPrefix(
			"/"+folder+"/",
			http.FileServer(http.Dir(filepath.Join(config.PublicPath, folder))),
		),
	)
}

func registerHandler() {
	serveStatic("js")
	serveStatic("css")

	http.HandleFunc("/", root)
	http.Handle("/play", NewPlayerServer())
}

func bootstrap() {
	indexTmpl = template.Must(template.ParseFiles(filepath.Join(config.PublicPath, "index.html")))
}

func serve() {
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.ListenPort, nil))
}

func startServer() {
	loadConfig()
	bootstrap()
	warmUp()
	registerHandler()
	serve()
}
