package main

import (
	"fmt"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
)

func main() {

	g, err := semconv.ParseGroups()
	// fmt.Printf("%+v\n", g)
	fmt.Println(err)

	for id, group := range g {
		fmt.Printf("%s: %+v\n", id, group.Attributes)
	}
}
