package redbutton

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"math/rand"
)

type RoomInfoMessage struct {
	Name  string `json:"name"`
	Marks int    `json:"marks"`
}

// a single websocket connection, listening to room events
type RoomListener struct {
	ws     *websocket.Conn
	room   *Room
	events chan RoomInfoMessage
}

func NewRoomListener(ws *websocket.Conn, room *Room) *RoomListener {
	return &RoomListener{ws: ws, events: make(chan RoomInfoMessage)}
}

// notifies this room listener that there's a new event
func (this *RoomListener) newEvent(message RoomInfoMessage) {
	err := websocket.JSON.Send(this.ws, &message)
	if err != nil {
		log.Println("failed sending json: " + err.Error())
		this.ws.Close()
		return
	}
}

type Room struct {
	listeners map[*RoomListener]bool
}

func NewVotingRoom() *Room {
	return &Room{
		listeners: map[*RoomListener]bool{},
	}
}

// TODO: possible race condition, this gets called from new WS connections
func (this *Room) registerListener(listener *RoomListener) {
	this.listeners[listener] = true
}

func (this *Room) unregisterListener(listener *RoomListener) {
	delete(this.listeners, listener)
}

func (this *Room) broadcastMessage(message RoomInfoMessage) {
	for listener, _ := range this.listeners {
		go listener.newEvent(message)
	}
}

func (this *server) eventsHandler(ws *websocket.Conn) {
	defer ws.Close()

	room := this.room

	listener := NewRoomListener(ws, room)
	room.registerListener(listener)
	defer room.unregisterListener(listener)

	// read data to catch the eof
	func() {
		for {
			var data []byte
			err := websocket.JSON.Receive(ws, &data)
			if err != nil {
				return
			}
		}
	}()

	log.Println("quitting marks handler")
}

type server struct {
	Port  string `envconfig:"PORT" default:"8081"`
	UiDir string `envconfig:"UIDIR" default:"."`
	room  *Room
}

func Main() {
	s := server{}
	envconfig.Process("redbutton", &s)
	s.UiDir, _ = filepath.Abs(s.UiDir)
	s.room = NewVotingRoom()
	fmt.Printf("config: port %s, ui: %s", s.Port, s.UiDir)

	msg := RoomInfoMessage{Name: "Development model meeting", Marks: 20}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.room.broadcastMessage(msg)
				msg.Marks = rand.Intn(3)
			}
		}
	}()

	http.Handle("/events", websocket.Handler(s.eventsHandler))
	http.Handle("/", http.FileServer(http.Dir(s.UiDir)))
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		panic(err.Error())
	}

}
