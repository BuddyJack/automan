package main

import (
	"github.com/go-redis/redis"
	"strings"
	"strconv"
	"unicode"
)

func main() {

	client := redis.NewClient(&redis.Options{Addr: "192.168.102.76:16379", DB: 0})
	res, _ := client.Info().Result()
	defer client.Close()
	items := strings.Split(res, "\n")
	for _, aa := range items {
		aa = strings.TrimFunc(aa, unicode.IsSpace)
		if strings.HasPrefix(aa, "#") || 0 == len(aa) {
			continue
		}
		fields := strings.Split(aa, ":")
		switch fields[0] {
		case "blocked_clients":
			blockClients, err := strconv.ParseUint(fields[1], 10, 64)
			if nil != err {
				blockClients = 10
			}
			println(blockClients)
		}
	}

}
