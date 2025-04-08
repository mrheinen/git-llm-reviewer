package main

import "fmt"

func main() {
	// This is a test file with deliberately poor code
	var x = 5
	var result = ""
	for i := 0; i < x; i++ {
		result = result + "test" // Inefficient string concatenation
	}
	fmt.Println(result)
}
