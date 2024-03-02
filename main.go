package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"image/jpeg"
)

var (
	quality = flag.Int("q", 60, "image quality")
)

var (
	ErrInvalidFile = errors.New("file does not match filter")
)

func getFilesRecursively(p string, filterFile func(filename string) bool) ([]string, error) {
	files := []string{}

	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !filterFile(path) {
			return nil
		}

		if !info.IsDir() {
			files = append(files, path)
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func compressJPEG(src *os.File, dst *os.File) error {
	img, err := jpeg.Decode(src)
	if err != nil {
		return err
	}

	return jpeg.Encode(dst, img, &jpeg.Options{
		Quality: *quality,
	})
}

func getDestinationFilename(f *os.File) string {
	fileName := f.Name()
	cleanFileName := strings.TrimSuffix(fileName, path.Ext(fileName))

	return fmt.Sprintf("%s_min%s", cleanFileName, path.Ext(fileName))
}

func isJPEG(filename string) bool {
	ext := path.Ext(filename)
	return ext == ".jpeg" || ext == ".jpg"
}

func main() {
	flag.Parse()

	fp := flag.Arg(0)

	files, err := getFilesRecursively(fp, func(filename string) bool {
		if strings.Contains(filename, "_min") {
			return false
		}

		if !isJPEG(filename) {
			return false
		}

		return true
	})
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		f, err := os.OpenFile(file, os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		dst, err := os.OpenFile(getDestinationFilename(f), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		err = compressJPEG(f, dst)
		if err != nil {
			panic(err)
		}
	}
}
