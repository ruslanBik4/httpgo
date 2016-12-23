package main

//go:generate /Users/rus/go/bin/qtc -dir=views/templates

import (
	"./templates/forms"
	"fmt"
)

func main() {
	fmt.Printf("%s\n", forms.ShowForm("Foo"))
}
