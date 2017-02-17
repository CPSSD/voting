package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/utils"
)

func main() {
	fmt.Println("Hello voting test file")

	utils.GenerateKeyPair(512)

	a := 3
	p := &a
	b := *p

	fmt.Println(a, *p, b)

	a = 5

	fmt.Println(a, *p, b)

	*p = 4

	fmt.Println(a, *p, b)
}
