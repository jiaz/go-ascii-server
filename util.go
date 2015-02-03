package main

import "C"

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}

func checkRet(ret C.int, err error) error {
	if int(ret) == 0 {
		return nil
	} else {
		return err
	}
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

func LockFile(filePath string) (bool, error) {
	dir, file := filepath.Split(filePath)
	if dir == filePath {
		return false, errors.New("Cannot lock dirs")
	}
	if _, err := os.Stat(dir); err != nil {
		return false, err
	}
	lockFileName := file + ".lock"
	_, err := os.OpenFile(filepath.Join(dir, lockFileName), os.O_CREATE|os.O_EXCL, 0666)
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
