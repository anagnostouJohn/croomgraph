package main

import (
	"fmt"
	"strconv"
)

func main() {

	x := ""

	if z, err := strconv.ParseFloat(x, 32); err == nil {
		fmt.Println("dasdsad")
		fmt.Println(z, "ASAS")

	} else {
		fmt.Println(err)
	}

}
