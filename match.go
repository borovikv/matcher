package main

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

const inputSize = 300
const minBitDistance = 8 // hamming distance
const minLength = 5      // min common region to be validated (seconds)

type SearchResult struct {
	name       string
	start, end float64
}

type WorkPair struct {
	first, second string
}

// Analyses the input then writes the results in the global result variable
func analyse(pair WorkPair) (SearchResult, SearchResult) {
	audio1 := readInts(pair.first)
	audio2 := readInts(pair.second)

	p1start, p1end, p2start, p2end := searchIntro(audio1, audio2)

	r1 := SearchResult{name: pair.first, start: p1start, end: p1end}
	r2 := SearchResult{name: pair.second, start: p2start, end: p2end}

	return r1, r2
}

func readInts(fname string) (nums []int64) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(b), ",")
	// Assign cap to avoid resize on every append.
	nums = make([]int64, 0, len(lines))

	for _, l := range lines {
		// Empty line occurs at the end of the file when we use Split.
		if len(l) == 0 {
			continue
		}
		n, err := strconv.ParseInt(strings.Replace(l, "\n", "", -1), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		nums = append(nums, n)
	}

	return nums
}

func searchIntro(fprint1 []int64, fprint2 []int64) (float64, float64, float64, float64) {
	// If lenghts/2 are not whole numbers trim them down a bit
	if len(fprint1)%2 != 0 {
		fprint1 = fprint1[0 : len(fprint1)-1]
		fprint2 = fprint2[0 : len(fprint2)-1]
	}

	// Get the offset with best score
	offset := getBestOffset(fprint1, fprint2)
	f1, f2 := getAllingedFingerprints(offset, fprint1, fprint2)
	hammed := hammItUp(f1, f2)

	// Find the contigious region
	start, end := findContiguousRegion(hammed, minBitDistance)
	if start < 0 || end < 0 {
		return 0.0, 0.0, 0.0, 0.0
	}

	//Convert everything to seconds
	secondsPerSample := float64(inputSize) / float64(len(fprint1))
	offsetInSeconds := float64(offset) * secondsPerSample
	commonRegionStart := float64(start) * secondsPerSample
	commonRegionEnd := float64(end) * secondsPerSample

	firstFileRegionStart := 0.0
	firstFileRegionEnd := 0.0

	secondFileRegionStart := 0.0
	secondFileRegionEnd := 0.0

	if offset >= 0 {
		firstFileRegionStart = commonRegionStart + offsetInSeconds
		firstFileRegionEnd = commonRegionEnd + offsetInSeconds

		secondFileRegionStart = commonRegionStart
		secondFileRegionEnd = commonRegionEnd
	} else {
		firstFileRegionStart = commonRegionStart
		firstFileRegionEnd = commonRegionEnd

		secondFileRegionStart = commonRegionStart - offsetInSeconds
		secondFileRegionEnd = commonRegionEnd - offsetInSeconds
	}

	// Check if the found region is bigger than min length
	if firstFileRegionEnd-firstFileRegionStart < minLength {
		return -1.0, -1.0, -1.0, -1.0
	}

	return firstFileRegionStart, firstFileRegionEnd, secondFileRegionStart, secondFileRegionEnd
}

// Returns the offset at wich audio aligns the best
// Value is positive if first audio is late
func getBestOffset(f1 []int64, f2 []int64) int {
	N := len(f1)
	endOfArray := N - 1
	iterations := N + 1 //one for the middle ground, 0 index

	diff := N/2 - 1

	a := N / 2
	b := endOfArray
	x := 0
	y := N/2 - 1
	upper := abs(a - b)

	output := make([]float64, iterations)

	for i := 0; i < iterations; i++ {
		output[i] = Compare(f1[a:a+upper], f2[x:x+upper])

		a = clip(a-1, 0, endOfArray)

		bVal := func() int {
			if diff < 0 {
				return b - 1
			}
			return b
		}
		b = clip(bVal(), 0, endOfArray)

		xVal := func() int {
			if diff < 0 {
				return x + 1
			}
			return x
		}
		x = clip(xVal(), 0, endOfArray)

		yVal := func() int {
			if diff >= 0 {
				return y + 1
			}
			return y
		}
		y = clip(yVal(), 0, endOfArray)

		diff--
	}

	index := getTheBiggestIndex(output)
	return (iterations-1)/2 - index
}

//Returns the trimmed arrays so the fingerprints data lines up
func getAllingedFingerprints(offset int, f1 []int64, f2 []int64) ([]int64, []int64) {
	if offset >= 0 {
		offsetCorrectedF1 := f1[offset:]
		offsetCorrectedF2 := f2[0 : len(f2)-offset]
		return offsetCorrectedF1, offsetCorrectedF2
	}

	offsetCorrectedF1 := f1[0 : len(f1)-abs(offset)]
	offsetCorrectedF2 := f2[abs(offset):]
	return offsetCorrectedF1, offsetCorrectedF2
}

func hammItUp(f1 []int64, f2 []int64) []int {
	result := make([]int, len(f1))
	for i := 0; i < len(f1); i++ {
		result[i] = CountBitsUint64(f1[i] ^ f2[i])
	}

	return result
}

func findContiguousRegion(arr []int, upperLimit int) (int, int) {
	start := -1
	end := -1

	for i := 0; i < len(arr); i++ {
		if arr[i] < upperLimit && nextOnesAreAlsoSmall(arr, i, upperLimit) {
			if start == -1 {
				start = i
			}
			end = i
		}
	}

	return start, end
}

func nextOnesAreAlsoSmall(arr []int, index int, upperLimit int) bool {
	if index+3 < len(arr) {
		v1 := arr[index+1]
		v2 := arr[index+2]
		v3 := arr[index+3]
		average := (v1 + v2 + v3) / 3

		if average < upperLimit {
			return true
		}

		return false
	}

	return false
}

const (
	bitsperint = 64
)

func Compare(fprint1, fprint2 []int64) float64 {
	dist := 0

	for i, sub := range fprint1 {
		dist += myHamming(sub, fprint2[i])
	}

	score := 1 - float64(dist)/float64(len(fprint1)*bitsperint)
	return score
}

func myHamming(a, b int64) (dist int) {
	dist = strings.Count(strconv.FormatInt(int64(a^b), 2), "1")
	return dist
}

const (
	m1q uint64 = 0x5555555555555555
	m2q        = 0x3333333333333333
	m4q        = 0x0f0f0f0f0f0f0f0f
	hq         = 0x0101010101010101
)

func CountBitsUint64(e int64) int {
	x := uint64(e)
	// put count of each 2 bits into those 2 bits
	x -= (x >> 1) & m1q

	// put count of each 4 bits into those 4 bits
	x = (x & m2q) + ((x >> 2) & m2q)

	// put count of each 8 bits into those 8 bits
	x = (x + (x >> 4)) & m4q

	// returns left 8 bits of x + (x<<8) + (x<<16) + (x<<24) + ...
	return int((x * hq) >> 56)
}
