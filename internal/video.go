package internal

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

func GetResolution(videoPath string) (int, int, error) {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-hide_banner", "-f", "null", "-")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Println(stderr.String())
		return 0, 0, err
	}

	re := regexp.MustCompile(`, (\d{2,5})x(\d{2,5})`)
	matches := re.FindStringSubmatch(stderr.String())
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("failed to get resolution")
	}
	width, _ := strconv.Atoi(matches[1])
	height, _ := strconv.Atoi(matches[2])
	return width, height, nil
}

func ExtractFrames(videoPath, outputDir string, rw, rh, fps int, cmdStrChan chan string) error {
	scale := strconv.Itoa(rw) + "x" + strconv.Itoa(rh)
	fpsStr := strconv.Itoa(fps)
	// ffmpeg -i wl.mp4 -f lavfi -i color=gray:s=40x30 -f lavfi -i color=black:s=40x30 -f lavfi -i color=white:s=40x30 -filter_complex "[0:v]scale=40x30,fps=12,threshold" wk/frame%04d.png
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-f", "lavfi", "-i", "color=gray:s="+scale,
		"-f", "lavfi", "-i", "color=black:s="+scale, "-f", "lavfi", "-i", "color=white:s="+scale,
		"-filter_complex", "[0:v]scale="+scale+",fps="+fpsStr+",threshold", outputDir+"/frame%04d.png")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmdStrChan <- cmd.String()
	close(cmdStrChan)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ReadImageAsBoolArray(filePath string) ([][]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	boolArray := make([][]bool, height)
	for y := 0; y < height; y++ {
		boolArray[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0xffff && g == 0xffff && b == 0xffff {
				boolArray[y][x] = true
			} else {
				boolArray[y][x] = false
			}
		}
	}

	return boolArray, nil
}
