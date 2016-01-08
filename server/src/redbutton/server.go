package redbutton

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
	"path/filepath"
)

type RoomStatusChangeEvent struct {
	RoomName     string `json:"name"`
	NumFlags     int    `json:"marks"`
	NumListeners int    `json:"listeners"`
}

// an entity that is interested in room events
// receives notifications via provided websocket connection
type RoomListener struct {
	ws     *websocket.Conn
	room   *Room
	events chan RoomStatusChangeEvent
}

func NewRoomListener(ws *websocket.Conn, room *Room) *RoomListener {
	return &RoomListener{
		ws:     ws,
		room:   room,
		events: make(chan RoomStatusChangeEvent),
	}
}

// notifies this room listener that there's a new event
func (this *RoomListener) newEvent(message interface{}) {
	err := websocket.WriteJSON(this.ws, message)
	if err != nil {
		log.Println("failed sending json: " + err.Error())
		this.ws.Close()
		return
	}
}

type Room struct {
	name         string
	listeners    map[*RoomListener]bool
	unhappyVotes map[string]bool
}

func NewVotingRoom() *Room {
	return &Room{
		listeners:    map[*RoomListener]bool{},
		unhappyVotes: map[string]bool{},
	}
}

// TODO: possible race condition, this gets called from new WS connections
func (this *Room) registerListener(listener *RoomListener) {
	this.listeners[listener] = true
	this.notifyStatusChanged()
}

func (this *Room) unregisterListener(listener *RoomListener) {
	delete(this.listeners, listener)
	this.notifyStatusChanged()
}

// builds and sends out a RoomStatusChangeEvent to this room
func (this *Room) notifyStatusChanged() {
	this.broadcastMessage(RoomStatusChangeEvent{
		RoomName:     this.name,
		NumFlags:     len(this.unhappyVotes),
		NumListeners: len(this.listeners),
	})
}

// broadcasts a message to all room listeners
func (this *Room) broadcastMessage(message interface{}) {
	for listener, _ := range this.listeners {
		go listener.newEvent(message)
	}
}

func (this *Room) setHappy(voterId string, happy bool) {
	if happy {
		delete(this.unhappyVotes, voterId)
	} else {
		this.unhappyVotes[voterId] = happy
	}
	this.notifyStatusChanged()
}

type server struct {
	Port  string `envconfig:"PORT" default:"8081"`
	UiDir string `envconfig:"UIDIR" default:"."`
	room  *Room
}

var websocketUpgrader = websocket.Upgrader{}

/**
this handler reports room events into provided websocket connection
*/
func (this *server) roomEventListenerHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	room := this.room

	listener := NewRoomListener(ws, room)
	room.registerListener(listener)
	defer room.unregisterListener(listener)

	// read data to catch the eof
	for {
		var data []byte
		err := websocket.ReadJSON(ws, &data)
		if err != nil {
			return
		}
	}
}

type VoteRequest struct {
	Happy bool `json:"happy"`
}

// voter ID comes from request
func (this *server) handleVote(request VoteRequest, r render.Render, params martini.Params, sess sessions.Session) {
	room := this.room // TODO: replace with room lookup

	voterId := sess.Get("voterId").(string)
	room.setHappy(voterId, request.Happy)
	println("oh yeah", voterId, request.Happy)
	r.JSON(200, &request)
}

func Main() {
	s := server{}
	envconfig.Process("redbutton", &s)
	s.UiDir, _ = filepath.Abs(s.UiDir)
	s.room = NewVotingRoom()
	s.room.name = "Very Important Meeting"
	fmt.Printf("config: port %s, ui: %s", s.Port, s.UiDir)

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte("it's not really a secret"))
	m.Use(sessions.Sessions("redbutton",store))
	m.Use(func(ses sessions.Session){
		id := ses.Get("voterId")
		if id!=nil {
			return
		}

		ses.Set("voterId",voterId())
	})
	m.Use(render.Renderer())
	m.Use(martini.Static(s.UiDir, martini.StaticOptions{Prefix: ""}))
	m.Get("/events", s.roomEventListenerHandler)
	m.Post("/vote/", binding.Bind(VoteRequest{}), s.handleVote)
	m.RunOnAddr("0.0.0.0:" + s.Port)

}
