package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	allFiles := listAllFiles()
	pairs := pairUpFiles(allFiles)
	for _, pair := range pairs {
		p1, p2 := analyse(pair)
		fmt.Println(p1, p2)
	}
}

/*
read all waf files
return [Jojo-01.wav Jojo-02.wav Jojo-03.wav Jojo-04.wav Jojo-05.wav Jojo-06.wav Jojo-07.wav Jojo-08.wav Jojo-09.wav Jojo-10.wav].........
*/
func listAllFiles() []string {
	dir := "./fingerprints/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = fmt.Sprintf(dir+"%s", f.Name())
	}
	return fileNames
}

//func pairUpFiles(singles []string) []WorkPair {
//	results := make([]WorkPair, len(singles)-1)
//
//	for i := 0; i < len(singles)-1; i++ {
//		results[i] = WorkPair{first: singles[i], second: singles[i+1]}
//	}
//
//	return results
//}

func pairUpFiles(files []string) []WorkPair {
	results := make([]WorkPair, len(files)*(len(files)+1)/2-4)
	k := 0
	for i := 0; i < len(files); i++ {
		for j := i; j < len(files); j++ {
			if i != j {
				results[k] = WorkPair{first: files[i], second: files[j]}
				k++
			}
		}
	}
	return results
}
