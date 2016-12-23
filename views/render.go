package main

//go:generate /Users/rus/go/bin/qtc -dir=views/templates

import (
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"fmt"
)

func main() {
	fmt.Printf("%s\n", forms.ShowForm("Foo"))
}
