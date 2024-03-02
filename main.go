package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"image/jpeg"
)

var (
	quality = flag.Int("q", 60, "image quality")
	rewrite = flag.Bool("r", false, "rewrite original files")
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

func compressJPEG(src io.Reader, dst io.Writer) error {
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
	flag.Usage = func() {
		fmt.Println("Usage: compress [flags] [path]")
		flag.PrintDefaults()
	}

	flag.Parse()

	fp := flag.Arg(0)
	if fp == "" {
		flag.Usage()
		os.Exit(1)
	}

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

		data, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		src := bytes.NewReader(data)

		contentType := http.DetectContentType(data)
		if contentType != "image/jpeg" {
			continue
		}

		dstFileName := getDestinationFilename(f)
		if *rewrite {
			dstFileName = f.Name()
		}

		dst, err := os.OpenFile(dstFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		err = compressJPEG(src, dst)
		if err != nil {
			panic(err)
		}
	}
}
