package main

import "anti-brute-force/internal/server"

func main() {
	s := server.NewServer()
	err := s.Start()
	if err != nil {
		return
	}
}
