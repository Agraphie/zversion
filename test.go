package main

import (
	"log"
)

func main() {
	//client := &http.Client{}
	//
	//req, err := http.NewRequest("HEAD", "http://137.226.113.4", nil)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	////req.Header.Set("User-Agent", "Curl/7.1.20")
	//
	//resp, err := client.Do(req)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	////defer resp.Body.Close()
	////body, err := ioutil.ReadAll(resp.Body)
	////if err != nil {
	////	log.Fatalln(err)
	////}
	//
	//log.Println(resp.Header)

	//req1, err1 := http.Get("http://69.164.201.157")
	//req1.Body.Close()
	//if err1 != nil {
	//	log.Fatalln(err1)
	//}
	//log.Println(req1.Header["Server"])

	if "1.1.03" < "1.1.18" {
		log.Println("Good!")
	}
}
