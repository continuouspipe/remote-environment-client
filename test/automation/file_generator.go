// In order to test a long running 'watch' task process
// This tool creates n files, in each x seconds interval
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const INTERVAL_IN_SECONDS = 4
const FILES_CREATED = 500

func main() {
	ticker := time.NewTicker(time.Second * INTERVAL_IN_SECONDS)

	c := 0

	for t := range ticker.C {
		for i := 0; i < FILES_CREATED; i = i + 1 {
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
		c = c + FILES_CREATED
		fmt.Printf("\n\nCreated %s files: %s. Total file created %s\n", strconv.Itoa(FILES_CREATED), t.String(), strconv.Itoa(c))
	}
}
