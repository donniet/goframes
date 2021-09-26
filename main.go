package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	civ := Skyciv{}
	if b, err := json.MarshalIndent(civ, "", "\t"); err != nil {
		fmt.Printf("error: %e", err)
	} else {
		os.Stdout.Write(b)
	}
}
