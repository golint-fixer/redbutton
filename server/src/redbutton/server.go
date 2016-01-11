package redbutton

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/martini-contrib/binding"
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

type ServerConfig struct {
	Port  string `envconfig:"PORT" default:"8081"`
	UiDir string `envconfig:"UIDIR" default:"."`
}

type server struct {
	ServerConfig
	room *Room
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

type VoterStatus struct {
	Happy bool   `json:"happy"`
}

// voter ID comes from request
func (this *server) handleChangeVoterStatus(request VoterStatus, r render.Render, params martini.Params, sess sessions.Session) {
	voterId := params["voterId"]
	if (voterId == "") {
		r.JSON(400, nil) // TODO: better error handling
		return
	}
	room := this.room // TODO: replace with room lookup

	room.setHappy(voterId, request.Happy)
	this.getVoterStatus(params, r)
}

func (this *server) getVoterStatus(params martini.Params, r render.Render) {
	result := VoterStatus{}
	result.Happy = true
	voterId := params["voterId"]
	if _, ok := this.room.unhappyVotes[voterId]; ok {
		result.Happy = false
	}
	r.JSON(200, &result)
}

type LoginResponse struct {
	VoterId string `json:"voterId"`
}

// so far only a stub of login service; returns new voterId each time it's called
func handleLogin(r render.Render) {
	response := LoginResponse{
		VoterId: voterId(),
	}
	r.JSON(200, &response)
}

func runServer(config ServerConfig) {
	s := server{ServerConfig:config}
	s.room = NewVotingRoom()
	s.room.name = "Very Important Meeting"

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte("it's not really a secret"))
	m.Use(sessions.Sessions("redbutton", store))
	m.Use(func(ses sessions.Session) {
		id := ses.Get("voterId")
		if id != nil {
			return
		}

		ses.Set("voterId", voterId())
	})
	m.Use(render.Renderer())
	m.Get("/events", s.roomEventListenerHandler)
	m.Post("/voter/:voterId", binding.Bind(VoterStatus{}), s.handleChangeVoterStatus)
	m.Get("/voter/:voterId", s.getVoterStatus)
	m.Post("/login", handleLogin)
	m.Use(martini.Static(s.UiDir, martini.StaticOptions{Prefix: ""}))
	m.RunOnAddr("0.0.0.0:" + s.Port)
}

func Main() {
	config := ServerConfig{}
	envconfig.Process("redbutton", &config)
	config.UiDir, _ = filepath.Abs(config.UiDir)

	fmt.Printf("config: port %s, ui: %s", config.Port, config.UiDir)

	runServer(config)
}
