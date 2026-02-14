package initializer

import (
	"fmt"
	"time"
)

type File struct {
	Name      string
	extension string
	hash      string
	delta     time.Duration
	changes   []change
}

type change struct {
	before string
	after  string
}

func Init() {
	fmt.Println("Hello, World!")
}
