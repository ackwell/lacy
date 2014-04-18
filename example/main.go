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

func sayHello(bot *lacy.Lacy, event lacy.Event) {
	bot.Out <- "Your friendly neighbourhood Lacybot says hello."
}

func main() {
	bot, err := lacy.New(lacy.Settings{
		Username: "bot",
		Password: "password",
		Origin: "http://localhost:8080",
		Plaintext: true,
	})
	if err != nil {
		panic(err) // programming
	}
	bot.Register([]string{"*"}, echo)
	bot.Register([]string{"mesg:out"}, yellBack)
	bot.Register([]string{"auth:succeed"}, sayHello)
	go bot.Run()
	<-bot.Quit
}
