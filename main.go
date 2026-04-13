package main

import (
	"context"
	"log"
	"os"

	"qml-language-server/handler"
)

func main() {
	h := handler.New(nil)
	if err := h.Serve(context.Background()); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
