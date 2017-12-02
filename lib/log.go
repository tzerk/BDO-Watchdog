package lib

import (
	"os"
	"log"
)

func logger(msg string) (err error) {

	// create your file with desired read/write permissions
	f, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}

	// defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	// set output of logs to f
	log.SetOutput(f)

	// write message
	log.Println(msg)

	return err
}