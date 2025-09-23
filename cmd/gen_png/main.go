package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x10, G: 0x90, B: 0xE0, A: 0xFF})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
	f, err := os.CreateTemp("", "upload-test-*.png")
	if err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
	fmt.Print(f.Name())
}
