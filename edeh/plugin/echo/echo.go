package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println("running echo plugin in ", wd)
	scn := bufio.NewScanner(os.Stdin)
	for scn.Scan() {
		fmt.Printf("Echo:[%s]\n", scn.Text())
	}
}
