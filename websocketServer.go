package main

import (
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func LockFile(filePath string) (bool, error) {
	dir, file := filepath.Split(filePath)
	if dir == filePath {
		return false, errors.New("Cannot lock dirs")
	}
	if _, err := os.Stat(dir); err != nil {
		return false, err
	}
	lockFileName := file + ".lock"
	_, err := os.Create(filepath.Join(dir, lockFileName))
	if err != nil {
		return false, err
	}
	return true, nil
}

func UnlockFile(filePath string) (bool, error) {
	dir, file := filepath.Split(filePath)
	if dir == filePath {
		return false, errors.New("Cannot unlock dirs")
	}
	if _, err := os.Stat(dir); err != nil {
		return false, err
	}
	lockFileName := file + ".lock"
	err := os.Remove(filepath.Join(dir, lockFileName))
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func checkFileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

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
			data.VideoBuffer[data.FrameCount], err = converter.ConvertToHtml(image)
			if err != nil {
				fatal(err)
			}

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

func registerHandler() {
	ws := websocket.Server{
		Handler: func(conn *websocket.Conn) {
			log.Println("Start streaming")
			for _, frame := range data.VideoBuffer[0:data.FrameCount] {
				websocket.Message.Send(conn, frame)
			}
			log.Println("Finished streaming")
		}}

	http.HandleFunc("/", root)
	http.Handle("/play", ws)
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
