package main

import (
	"fmt"
	"helm.sh/helm/v3/pkg/action"
)

func main() {
	fmt.Println("HALLOJ cmd/helm3-agent/main.go")
	fmt.Printf("HALLOJ %v\n", new(action.Configuration))
}
