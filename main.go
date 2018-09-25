package main

import (
	"hlcup_epoll/server"
	)

func main() {

	epollServer := server.NewServer(
		8080,
		"/home/artyomnorin/projects/go/src/hlcup_epoll/data/full/data.zip",
		"/home/artyomnorin/projects/go/src/hlcup_epoll/data/full/options.txt")

	epollServer.Run()
}
