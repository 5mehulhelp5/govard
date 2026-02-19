//go:build !desktop

package main

import "fmt"

func main() {
	fmt.Println("Govard Desktop is not built yet. Run `govard desktop --dev` or build with Wails.")
}
