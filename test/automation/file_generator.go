// In order to test a long running 'watch' task process
// This tool creates 50 files each 15 seconds
package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
)

func main() {
	ticker := time.NewTicker(time.Second * 15)

	c := 0

	for t := range ticker.C {
		for i := 0; i < 50; i = i + 1 {
			now := time.Now()
			fileName := "file_" + now.Format(time.RFC3339Nano) + "_.txt"
			file, err := os.Create(fileName)
			if err != nil {
				fmt.Println(err.Error())
			}
			file.WriteString("Tick at " + t.String())
			file.Close()
			fmt.Printf("Created file %s\n", fileName)
		}
		c = c + 50
		fmt.Println("\nCreated 50 files: " + t.String() + ". Total file created " + strconv.Itoa(c))
	}
}
