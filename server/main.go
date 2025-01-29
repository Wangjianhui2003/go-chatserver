package main

import server2 "JHServer/server"

func main() {
	server := server2.NewServer("127.0.0.1", "8001")
	server.Start()
}
