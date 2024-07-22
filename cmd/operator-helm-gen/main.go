package main

import (
	"time"
)

func main() {
	println(time.Now().UTC().Format(time.RFC3339))
}
