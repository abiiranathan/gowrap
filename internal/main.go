package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/abiiranathan/gowrap/server"
	"github.com/abiiranathan/gowrap/ws"
)

type User struct {
	ID   uint
	Name string
	Age  int
}

func setupClient() {
	dialer := ws.NewDialer("ws://localhost:8080/ws")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	u := User{1, "Abiira", 30}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("we are done sending messages..")
			return
		default:
			time.Sleep(time.Second * 5)
			err := dialer.Send(u)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func main() {
	hub, quit := ws.NewHandler()
	defer quit()

	go hub.Run()
	go setupClient()

	mux := http.DefaultServeMux
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/ws", hub)

	svr := server.NewServer(":8080", server.WithHandler(mux))
	svr.Run()
}
