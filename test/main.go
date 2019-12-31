package main

import (
	"encoding/json"
	"fmt"
	"github.com/innodv/errors"
)

func foo() {
	if 1 == 1 {
		bar()
	}

}

func bar() {
	if 2+3 == 5 {
		baz()
	}

}

func baz() {
	err := errors.New("fo")
	fmt.Println(err)
	res, _ := json.Marshal(err)
	fmt.Println(string(res))
}

func main() {
	foo()
}
