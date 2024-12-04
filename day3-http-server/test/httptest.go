package main

import (
	"day3-http-server"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geecache.NewGroup("score", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	addr1 := "localhost:9999"
	peers1 := geecache.NewHTTPPool(addr1)
	log.Println("geecache is running at", addr1)
	log.Fatal(http.ListenAndServe(addr1, peers1))

}
