package main

import (
	"ThisBot/components"
	"os"
	"time"
)

func main() {
	args := os.Args

	if len(args) >= 3 && args[1] == "-c" {
		old := args[2]
		_, err := os.Stat(old)
		if err == nil {
			for i := 0; i < 10; i++ {
				os.Remove(old)
				if err != nil || os.IsNotExist(err) {
					break
				}
				time.Sleep(time.Duration(1000) * time.Millisecond)
			}
		}
	}

	components.Run()

	for {
		time.Sleep(20 * time.Second)
	}
}
