package main

import "fmt"

func progressBar(percent int,length int) {
	fmt.Print("\r[")
	for i := range length {
		if i*100/length<=percent {
			fmt.Print("#")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Printf("] %d%%",percent)
}
