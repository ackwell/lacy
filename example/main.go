package main

import (
	"fmt"
	"github.com/sysr-q/lacy"
	"strings"
)

func echo(bot *lacy.Lacy, event lacy.Event) {
	fmt.Printf("%#v\n", event)
}

func yellBack(bot *lacy.Lacy, event lacy.Event) {
	payload := event.Payload.(map[string]string)
	bot.Out <- strings.ToUpper(payload["message"])
}

func main() {
	bot, err := lacy.New(lacy.Settings{"bot", "password", "http://localhost:8080"})
	if err != nil {
		panic(err) // programming
	}
	bot.Register([]string{"*"}, echo)
	bot.Register([]string{"mesg:out"}, yellBack)
	go bot.Run()
	<-bot.Quit
}