package main

import (
	"ThisBot/components"
	"time"
)

func main() {
	components.Run()

	for {
		time.Sleep(20 * time.Second)
	}
}
