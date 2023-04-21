package main

import (
	"fmt"
	socketIoClient "socket-io-client-go"
	"time"
)

func main() {
	nameSpace := "/server/realtime"
	client := socketIoClient.Client{NameSpace: &nameSpace}
	client.On("connect", func(client *socketIoClient.Client, data []string) {
		fmt.Println("connected")
		client.Emit("register", map[string]string{"key": "server-8fcb4ae3042ddf364d2a9fa16597b92dc3c22aa7"})
	})

	client.On("disconnect", func(client *socketIoClient.Client, data []string) {
		fmt.Println("disconnect")
	})

	client.On("test", func(client *socketIoClient.Client, data []string) {
		fmt.Println(data)
	})

	client.On("update", func(client *socketIoClient.Client, data []string) {
		fmt.Println("update", data)
	})

	if err := client.Connect("https://featureprobe.io/server/realtime", "polling"); err != nil {
		fmt.Println(err)
		return
	}

	for {
		time.Sleep(5 * time.Second)
	}
}
