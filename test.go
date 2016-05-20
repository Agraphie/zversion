package main

import (
	"net/http"
	"fmt"
	"net"
)

func main() {
	req, err := http.NewRequest("HEAD", net.ParseIP("80.14.249.95"), nil)
	client := http.Client{}
	resp, err := client.Do(req)
	//resp, err := http.Head("http://80.14.249.95")
	if err != nil {
		// handle error
	}

	fmt.Println(resp)

}
