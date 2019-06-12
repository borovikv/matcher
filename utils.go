package main

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clip(val int, min int, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func getTheBiggestIndex(arr []float64) int {
	index := 0
	biggest := 0.0

	for i := 0; i < len(arr); i++ {
		val := arr[i]
		if val > biggest {
			biggest = val
			index = i
		}
	}

	return index
}
