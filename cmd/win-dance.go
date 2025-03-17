package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/homearchbishop/win-dance/internal"
	flag "github.com/spf13/pflag"
)

var (
	startTime   time.Time
	elapsedTime time.Duration
)

func main() {
	info := color.New(color.FgCyan).SprintFunc()
	item := color.New(color.FgBlue).SprintFunc()
	sub := color.New(color.FgHiBlack).SprintFunc()
	warning := color.New(color.FgYellow).SprintFunc()
	errormsg := color.New(color.FgRed).SprintFunc()
	success := color.New(color.FgGreen).SprintFunc()

	// Parse cli args
	outputFileP := flag.StringP("output", "o", "", "output file")
	workDirP := flag.StringP("workdir", "w", "windance-working-directory", "working directory")
	fpsP := flag.IntP("fps", "f", 12, "frames per second")
	resolutionVerticalP := flag.IntP("resolution-vertical", "r", 30, "resolution")
	keepWorkdirP := flag.BoolP("keep-workdir", "k", false, "keep working directory")
	flag.Parse()
	outputFile := *outputFileP
	workDir := *workDirP
	workDirFrameDir := workDir + "/frames"
	fps := *fpsP
	resolutionVertical := *resolutionVerticalP
	keepWorkdir := *keepWorkdirP
	inputFile := flag.Args()[0]
	if outputFile == "" {
		fmt.Println(errormsg("output file is required, use -o or --output"))
		return
	}
	if !strings.HasSuffix(outputFile, ".exe") {
		outputFile = outputFile + ".exe"
	}

	fmt.Println(info("[Using cli args]"))
	fmt.Println("📄 output file:", sub(outputFile))
	fmt.Println("📁 working directory:", sub(workDir))
	fmt.Println("📽️  input video:", sub(inputFile))
	fmt.Println("🎞️  fps:", sub(fps))
	fmt.Println("🫧  resolution (vertical):", sub(resolutionVertical), sub("px"))
	fmt.Println("🧹 keep working directory:", sub(keepWorkdir))

	// Check if the input file exist & working directory empty
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		os.Mkdir(workDir, 0755)
		os.Mkdir(workDirFrameDir, 0755)
	} else {
		files, err := os.ReadDir(workDir)
		if err != nil {
			fmt.Printf("%s failed to read working directory: %s\n", errormsg("Error:"), workDir)
			return
		}
		if len(files) != 0 {
			fmt.Printf("%s working directory is not empty: %s, Overwrite? [y/n]: ", warning("Warning:"), sub(workDir))
			var input string
			fmt.Scanln(&input)
			if input != "y" {
				fmt.Println(warning("Abort"))
				return
			}
			os.RemoveAll(workDir)
			os.Mkdir(workDir, 0755)
		}
		os.Mkdir(workDirFrameDir, 0755)
	}
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("%s input file does not exist: %s\n", errormsg("Error:"), inputFile)
		return
	}

	fmt.Println(info("[Resolve video] "))

	// Handle the input video: calculate resolution
	fmt.Printf("%s ", item("Get resolution..."))
	startTime = time.Now()
	videoWidth, videoHeight, err := internal.GetResolution(inputFile)
	if err != nil {
		fmt.Println(errormsg("failed to get video resolution"))
		fmt.Println(err)
		return
	}
	frameWidth := int(float64(resolutionVertical) * float64(videoWidth) / float64(videoHeight))
	frameHeight := resolutionVertical
	elapsedTime = time.Since(startTime)
	fmt.Printf("%dx%d -> %dx%d (%d ms)\n", videoWidth, videoHeight, frameWidth, frameHeight, elapsedTime.Milliseconds())

	// Handle the input video: extract frames
	fmt.Printf("%s ", item("Extract frames..."))
	extCmdStrChan := make(chan string, 1)
	extFlagChan := make(chan bool, 1)
	startTime = time.Now()
	go func() {
		err = internal.ExtractFrames(inputFile, workDirFrameDir, frameWidth, frameHeight, fps, extCmdStrChan)
		if err != nil {
			fmt.Println(errormsg("failed to extract frames"))
			fmt.Println(err)
			extFlagChan <- false
			close(extFlagChan)
			return
		}
		extFlagChan <- true
		close(extFlagChan)
	}()
	cmdStr := <-extCmdStrChan
	fmt.Print(sub(cmdStr), " ")
	extFlag := <-extFlagChan
	if !extFlag {
		return
	}
	framesImgFiles, err := os.ReadDir(workDirFrameDir)
	if err != nil {
		fmt.Println(errormsg("\n failed to read frames directory"))
		fmt.Println(err)
		return
	}
	elapsedTime = time.Since(startTime)
	fmt.Printf("totally %d frames (%d ms)\n", len(framesImgFiles), elapsedTime.Milliseconds())

	// Handle each frame from files
	fmt.Printf("%s ", item("Generate Rects of all frames..."))
	rectsQueue := make([]*[]*internal.Rect, len(framesImgFiles))
	var wg sync.WaitGroup
	var rqMux sync.Mutex
	fileIdxChan := make(chan int, len(framesImgFiles))
	maxWndCnt := 0
	for i := range framesImgFiles {
		fileIdxChan <- i
	}
	close(fileIdxChan)
	startTime = time.Now()
	for range 10 { // use 10 goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range fileIdxChan {
				// Handle the input frames: read as array
				file := framesImgFiles[i]
				frame, err := internal.ReadImageAsBoolArray(workDirFrameDir + "/" + file.Name())
				if err != nil {
					fmt.Printf("%s failed to read frame %s\n", errormsg("Error:"), file.Name())
					fmt.Println(err)
					return
				}
				// Do the algorithm
				rects := internal.CalcRectsOfFrame(frame, frameWidth, frameHeight)
				if len(*rects) > maxWndCnt {
					maxWndCnt = len(*rects)
				}
				// Save the result
				rqMux.Lock()
				rectsQueue[i] = rects
				rqMux.Unlock()
			}
		}()
	}
	wg.Wait()
	elapsedTime = time.Since(startTime)
	fmt.Printf("OK. Maximum %d window required (%d ms)\n", maxWndCnt, elapsedTime.Milliseconds())

	fmt.Println(info("[Create program]"))

	// Generate src code
	fmt.Printf("%s ", item("Generate src code..."))
	startTime = time.Now()
	src, err := internal.GenerateSrc(rectsQueue, fps, frameWidth, frameHeight)
	if err != nil {
		fmt.Println(err)
		return
	}
	elapsedTime = time.Since(startTime)
	fmt.Printf("OK (%d ms)\n", elapsedTime.Milliseconds())

	// Save src code to temp file
	fmt.Printf("%s ", item("Save src code to file..."))
	startTime = time.Now()
	srcFile, err := internal.SaveSrc(workDir, src)
	if err != nil {
		fmt.Println(err)
		return
	}
	elapsedTime = time.Since(startTime)
	fmt.Printf("%s OK (%d µs)\n", sub(srcFile), elapsedTime.Microseconds())

	// Compile temp file to output file
	fmt.Printf("%s ", item("Compile..."))
	cplCmdStrChan := make(chan string, 1)
	cplFlagChan := make(chan bool, 1)
	startTime = time.Now()
	go func() {
		err := internal.Compile(outputFile, srcFile, cplCmdStrChan)
		if err != nil {
			fmt.Println(errormsg("failed to compile"))
			fmt.Println(err)
			cplFlagChan <- false
			close(cplFlagChan)
			return
		}
		cplFlagChan <- true
		close(cplFlagChan)
	}()
	cmdStr = <-cplCmdStrChan
	fmt.Print(sub(cmdStr), " ")
	cplFlag := <-cplFlagChan
	if !cplFlag {
		return
	}
	elapsedTime = time.Since(startTime)
	fmt.Printf("OK (%d ms)\n", elapsedTime.Milliseconds())

	// Clean up
	if !keepWorkdir {
		fmt.Printf("%s ", item("Clean up..."))
		os.RemoveAll(workDir)
		fmt.Println("OK")
	}

	fmt.Println(success("Done"))
}
