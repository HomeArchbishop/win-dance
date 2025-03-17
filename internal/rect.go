package internal

import (
	"container/heap"
	"fmt"
)

var DEBUG = false

type RectInfo struct {
	I1        int
	J1        int
	I2        int
	J2        int
	UnusedSqr int
}

type Rect struct {
	X int
	Y int
	W int
	H int
}

func CalcRectsOfFrame(frame [][]bool, animationResolutionWidth, animationResolutionHeight int) *[]*Rect {
	frameCoverage := make([][]bool, animationResolutionHeight)
	frameUncoverageCnt := 0
	rectsArray := make([][]*RectInfo, animationResolutionWidth)
	rects := &RectHeap{}
	heap.Init(rects)
	results := make([]*Rect, 0)

	for i := animationResolutionHeight - 1; i >= 0; i-- {
		frameCoverage[i] = make([]bool, animationResolutionWidth)
		rectsArray[i] = make([]*RectInfo, animationResolutionWidth)
		for j := animationResolutionWidth - 1; j >= 0; j-- {
			if !frame[i][j] {
				frameCoverage[i][j] = true
				continue
			}
			frameUncoverageCnt++
			frameCoverage[i][j] = false
			// Find the largest rectangle with the top-left corner at (i, j)
			// and the bottom-right corner at (i2, j2)
			maxsqr, maxi2, maxj2 := 0, i, j
			tailJ := animationResolutionWidth - 1
			for i2 := i; i2 <= animationResolutionHeight-1 && frame[i2][j]; i2++ {
				for j2 := j; j2 <= tailJ; j2++ {
					if !frame[i2][j2] {
						tailJ = j2 - 1
						break
					}
					sqr := (i2 - i + 1) * (j2 - j + 1)
					if sqr > maxsqr {
						maxsqr = sqr
						maxi2 = i2
						maxj2 = j2
					}
				}
			}
			// Add the rectangle to the list
			info := &RectInfo{i, j, maxi2, maxj2, maxsqr}
			rectsArray[i][j] = info
			heap.Push(rects, info)
		}
	}

	/* DEBUG */
	if false {
		for i := range animationResolutionHeight {
			for j := range animationResolutionWidth {
				if !frame[i][j] {
					fmt.Printf("0\t")
				} else {
					fmt.Printf("%d\t", rectsArray[i][j].UnusedSqr)
				}
			}
			fmt.Println()
		}
	}

	for frameUncoverageCnt > 0 {
		// Find the largest rectangle
		largestRect := heap.Pop(rects).(*RectInfo)
		if DEBUG {
			fmt.Printf("(%d,%d)(%d,%d)%d ", largestRect.I1, largestRect.J1, largestRect.I2, largestRect.J2, largestRect.UnusedSqr)
		}
		results = append(results, &Rect{
			X: largestRect.J1,
			Y: largestRect.I1,
			W: largestRect.J2 - largestRect.J1 + 1,
			H: largestRect.I2 - largestRect.I1 + 1,
		})
		// Adjust each rect info
		for _, rect := range *rects {
			multipartLeft := max(rect.J1, largestRect.J1)
			multipartRight := min(rect.J2, largestRect.J2)
			multipartTop := max(rect.I1, largestRect.I1)
			multipartBottom := min(rect.I2, largestRect.I2)
			if multipartLeft > multipartRight || multipartTop > multipartBottom {
				continue
			}
			for i := multipartTop; i <= multipartBottom; i++ {
				for j := multipartLeft; j <= multipartRight; j++ {
					if !frameCoverage[i][j] {
						rect.UnusedSqr--
					}
				}
			}
		}
		// Update the frame coverage
		supposedUncoverageCnt := frameUncoverageCnt - largestRect.UnusedSqr
		for i := largestRect.I1; i <= largestRect.I2; i++ {
			for j := largestRect.J1; j <= largestRect.J2; j++ {
				if !frameCoverage[i][j] {
					frameCoverage[i][j] = true
					frameUncoverageCnt--
				}
			}
		}
		if frameUncoverageCnt != supposedUncoverageCnt {
			panic("frameUncoverageCnt != supposedUncoverageCnt")
		}
		// Rebuild the heap
		heap.Init(rects)
	}

	if DEBUG {
		fmt.Println("-----")
	}

	return &results
}
