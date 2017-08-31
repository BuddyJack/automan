package main

import "strings"

func main() {
	var test = make(map[string][2]string)

	if 0==strings.Compare(test["1"][0],"") {
		println("test")
	}

}
