package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	config := &ssh.ClientConfig{
		ClientVersion: "Go SSH Client v1337",
		User:          "jpillora",
		Auth:          []ssh.AuthMethod{ssh.Password("t0ps3cr3t")},
	}

	client, err := ssh.Dial("tcp", "localhost:2200", config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected")

	go open(client, "foo")

	time.Sleep(1 * time.Second)
	go open(client, "bar")

	time.Sleep(1 * time.Second)
	go open(client, "bazz")

	client.Wait()

	log.Printf("disconnected")
}

func open(client *ssh.Client, chanType string) {

	//open channel
	channel, requests, err := client.OpenChannel(chanType, s("%s extra data", chanType))
	if err != nil {
		log.Fatal(err)
	}

	//requests must be serviced
	go ssh.DiscardRequests(requests)

	//send data forever...
	n := 1
	for {
		_, err := channel.Write(s("#%d send data channel ", n))
		if err != nil {
			break
		}
		n++
		time.Sleep(3 * time.Second)
	}
}

func s(f string, args ...interface{}) []byte {
	return []byte(fmt.Sprintf(f, args...))
}
