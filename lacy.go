package lacy

import "github.com/googollee/go-socket.io"

type Event struct {
	Name    string
	Payload interface{} // ideally this would be a map[string]interface{}
}

type Plugin func(*Lacy, Event)

// Settings allows easy reuse of user/password/origin between bot sessions.
type Settings struct {
	Username, Password, Origin string
	Plaintext bool
}

type Lacy struct {
	settings Settings
	client   *socketio.Client
	UserId   string
	Quit     chan interface{}
	Out      chan string
	plugins  map[string][]Plugin
}

// plugins

// Register registers a new "plugin" to receive events from Wigslace.
func (bot *Lacy) Register(events []string, plugin Plugin) {
	// Special case; * -> every event.
	if len(events) == 1 && events[0] == "*" {
		events = []string{
			"connect",
			"conn:ready",
			"auth:fail",
			"auth:succeed",
			"user:join",
			"user:part",
			"user:active",
			"user:get",
			"user:data",
			"user:list",
			"mesg:scrollback",
			"mesg:out",
		}
	}

	for _, event := range events {
		if _, ok := bot.plugins[event]; ok {
			bot.plugins[event] = append(bot.plugins[event], plugin)
		} else {
			bot.plugins[event] = []Plugin{plugin}
		}
	}
}

// fan distributes an event across all of the currently registered plugins.
func (bot *Lacy) fan(event Event) {
	if plugins, ok := bot.plugins[event.Name]; ok {
		for _, plugin := range plugins {
			plugin(bot, event)
		}
	}
}

// socket.io handlers

func (bot *Lacy) connect(ns *socketio.NameSpace) {
	bot.client.Emit("auth:do", map[string]string{
		"username": bot.settings.Username,
		"password": bot.settings.Password,
	})
}

func (bot *Lacy) connReady(ns *socketio.NameSpace, id string) {
	bot.UserId = id
	if bot.settings.Plaintext {
		bot.client.Emit("opts:set", map[string]interface{}{
			"option": "plaintext",
			"value": true,
		})
	}
}

func (bot *Lacy) authFail(ns *socketio.NameSpace) {
	bot.fan(Event{"auth:fail", nil})
	bot.client.Emit("disconnect") // please work
}

func (bot *Lacy) message(ns *socketio.NameSpace, msg map[string]string) {
	if msg["user"] == bot.UserId {
		return
	}
	bot.fan(Event{"mesg:out", msg})
}

// poll repeatedly polls the outbound message channel, sending off messages
// to the server as they're received.
func (bot *Lacy) poll() {
	for {
		msg := <-bot.Out
		bot.client.Emit("mesg:in", msg)
	}
}

func (bot *Lacy) Run() {
	go bot.poll()
	bot.client.Run()
	bot.Quit <- nil
}

func New(settings Settings) (bot *Lacy, err error) {
	client, err := socketio.Dial(settings.Origin)
	if err != nil {
		return
	}
	bot = &Lacy{
		settings: settings,
		client:   client,
		UserId:   "",
		Quit:     make(chan interface{}),
		Out:      make(chan string),
		plugins:  make(map[string][]Plugin),
	}
	// We care about these, so we'll handle them explicitly:
	bot.client.On("connect", bot.connect)
	bot.client.On("mesg:out", bot.message)
	bot.client.On("conn:ready", bot.connReady)
	bot.client.On("auth:fail", bot.authFail)
	// Apologies for the boilerplate, my attempts to meta-program around it
	// were met with bizarre function pointer reuse.
	bot.client.On("auth:succeed", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"auth:succeed", payload})
	})
	bot.client.On("user:join", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:join", payload})
	})
	bot.client.On("user:part", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:part", payload})
	})
	bot.client.On("user:active", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:active", payload})
	})
	bot.client.On("user:get", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:get", payload})
	})
	bot.client.On("user:data", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:data", payload})
	})
	bot.client.On("user:list", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"user:list", payload})
	})
	bot.client.On("mesg:scrollback", func(ns *socketio.NameSpace, payload interface{}) {
		bot.fan(Event{"mesg:scrollback", payload})
	})
	return
}
