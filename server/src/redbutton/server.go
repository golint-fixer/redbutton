package redbutton

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"path/filepath"
	"redbutton/api"
	"strings"
)

// an entity that is interested in room events
// receives notifications via provided websocket connection
type RoomListener struct {
	ws     *websocket.Conn
	room   *Room
	events chan api.RoomStatusChangeEvent
}

func NewRoomListener(ws *websocket.Conn, room *Room) *RoomListener {
	return &RoomListener{
		ws:     ws,
		room:   room,
		events: make(chan api.RoomStatusChangeEvent),
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
	id           string
	name         string
	owner        string
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
	this.broadcastMessage(api.RoomStatusChangeEvent{
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
	rooms map[string]*Room
}

var websocketUpgrader = websocket.Upgrader{}

/**
this handler reports room events into provided websocket connection
*/
func (this *server) roomEventListenerHandler(params martini.Params, w http.ResponseWriter, r *http.Request, ren render.Render) {
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	room, err := this.lookupRoomFromRequest(params)
	if err != nil {
		return
	}

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

func (this *server) lookupRoomFromRequest(params martini.Params) (*Room, error) {
	roomId := params["roomId"]
	if roomId == "" {
		return nil, ApiError{}.badRequest("room ID not found")
	}

	if _, ok := this.rooms[roomId]; !ok {
		return nil, ApiError{}.notFound("room " + roomId + " was not found")
	}

	return this.rooms[roomId], nil
}

// voter ID comes from request
func (this *server) handleChangeVoterStatus(request api.VoterStatus, r render.Render, params martini.Params) {
	voterId := params["voterId"]
	if voterId == "" {
		r.JSON(400, nil) // TODO: better error handling
		return
	}
	room, err := this.lookupRoomFromRequest(params)
	if err != nil {
		respondWithError(err, r)
		return
	}

	room.setHappy(voterId, request.Happy)
	this.getVoterStatus(params, r)
}

func (this *server) createRoom(info api.RoomInfo, r render.Render) {
	id := uniqueId()
	log.Println("creating room", id)

	info.RoomName = strings.TrimSpace(info.RoomName)
	if info.RoomName=="" {
		respondWithError(ApiError{}.badRequest("Room name is missing"),r)
		return
	}


	room := NewVotingRoom()
	this.rooms[id] = room
	room.id = id
	room.owner = info.RoomOwner
	room.name = info.RoomName

	r.Status(http.StatusCreated)
	this.getRoomInfo(martini.Params{"roomId": id}, r)
}

func (this *server) getRoomInfo(params martini.Params, r render.Render) {
	room, err := this.lookupRoomFromRequest(params)
	if err != nil {
		respondWithError(err, r)
		return
	}

	info := api.RoomInfo{}
	info.Id = room.id
	info.RoomName = room.name
	info.RoomOwner = room.owner

	r.JSON(http.StatusOK, &info)
}

func (this *server) getVoterStatus(params martini.Params, r render.Render) {
	room, err := this.lookupRoomFromRequest(params)
	if err != nil {
		respondWithError(err, r)
		return
	}

	result := api.VoterStatus{}
	result.Happy = true
	voterId := params["voterId"]
	if _, ok := room.unhappyVotes[voterId]; ok {
		result.Happy = false
	}
	r.JSON(200, &result)
}

// so far only a stub of login service; returns new voterId each time it's called
func handleLogin(r render.Render) {
	response := api.LoginResponse{
		VoterId: uniqueId(),
	}
	r.JSON(200, &response)
}

func runServer(config ServerConfig) {
	s := server{ServerConfig: config, rooms: map[string]*Room{}}

	room := NewVotingRoom()
	room.name = "Very Important Meeting"
	s.rooms["default"] = room

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Group("/api",func(r martini.Router){
		r.Get("/events/:roomId", s.roomEventListenerHandler)
		r.Post("/room/:roomId/voter/:voterId", binding.Bind(api.VoterStatus{}), s.handleChangeVoterStatus)
		r.Get("/room/:roomId/voter/:voterId", s.getVoterStatus)
		r.Get("/room/:roomId", s.getRoomInfo)
		r.Post("/room", binding.Bind(api.RoomInfo{}), s.createRoom)
		r.Post("/login", handleLogin)
	})
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
