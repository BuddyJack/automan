package main

import (
	"os/exec"
)

func main() {

	cmd := exec.Command("sh", "-c", "ps aux|grep -v PID|sort -rn -k3|head -3")
	bs,err :=cmd.Output()
	if nil!=err{
		println("error")
	}else {
		println(string(bs))
	}
}
