package redbutton

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
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
func (this *server) roomEventListenerHandler(resp http.ResponseWriter, req *http.Request) {
	ws, err := websocketUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	c := api.NewHandler(req)
	room := this.lookupRoomFromRequest(c)
	if room == nil {
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

func (this *server) lookupRoomFromRequest(c *api.HttpHandlerContext) *Room {
	roomId := c.PathParam("roomId")
	if roomId == "" {
		c.Error(http.StatusBadRequest, "room ID not found")
		return nil
	}

	if _, ok := this.rooms[roomId]; !ok {
		c.Error(http.StatusNotFound, "room "+roomId+" was not found")
		return nil
	}

	return this.rooms[roomId]
}

// voter ID comes from request
func (this *server) handleChangeVoterStatus(c *api.HttpHandlerContext) {
	voterId := c.PathParam("voterId")
	if voterId == "" {
		c.Error(http.StatusBadRequest, "voter ID missing")
		return
	}
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	request := api.VoterStatus{}
	if !c.ParseRequest(&request) {
		return
	}

	room.setHappy(voterId, request.Happy)
	this.getVoterStatus(c)
}

func (this *server) createRoom(c *api.HttpHandlerContext) {
	id := uniqueId()
	log.Println("creating room", id)

	info := api.RoomInfo{}
	if !c.ParseRequest(&info) {
		return
	}
	info.RoomName = strings.TrimSpace(info.RoomName)
	if info.RoomName == "" {
		c.Error(http.StatusBadRequest, "Room name is missing")
		return
	}

	room := NewVotingRoom()
	this.rooms[id] = room
	room.id = id
	room.owner = info.RoomOwner
	room.name = info.RoomName

	c.Status(http.StatusCreated).Result(this.mapRoomInfo(room))
}

func (this *server) getRoomInfo(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	c.Result(this.mapRoomInfo(room))
}

func (this *server) mapRoomInfo(room *Room) *api.RoomInfo {
	info := api.RoomInfo{}
	info.Id = room.id
	info.RoomName = room.name
	info.RoomOwner = room.owner
	return &info
}

func (this *server) getVoterStatus(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	result := api.VoterStatus{}
	result.Happy = true
	voterId := c.PathParam("voterId")
	if _, ok := room.unhappyVotes[voterId]; ok {
		result.Happy = false
	}
	c.Result(&result)
}

// so far only a stub of login service; returns new voterId each time it's called

func (this *server) router() http.Handler {
	m := mux.NewRouter()

	r := api.NewRouteWrapper(m.PathPrefix("/api").Subrouter())

	r.Post("/login", func(c *api.HttpHandlerContext) {
		c.Result(&api.LoginResponse{
			VoterId: uniqueId(),
		})
	})

	r.Post("/room", this.createRoom)
	r.Get("/room/{roomId}", this.getRoomInfo)
	r.Get("/room/{roomId}/voter/{voterId}", this.getVoterStatus)
	r.Post("/room/{roomId}/voter/{voterId}", this.handleChangeVoterStatus)
	r.Router.Methods("GET").Path("/events/{roomId}").HandlerFunc(this.roomEventListenerHandler)
	m.PathPrefix("/").Handler(http.FileServer(http.Dir(this.UiDir)))

	return m
}

func runServer(config ServerConfig) {
	s := server{ServerConfig: config, rooms: map[string]*Room{}}

	room := NewVotingRoom()
	room.name = "Very Important Meeting"
	s.rooms["default"] = room

	http.Handle("/", s.router())
	http.ListenAndServe(":"+s.Port, nil)
}

func Main() {
	config := ServerConfig{}
	envconfig.Process("redbutton", &config)
	config.UiDir, _ = filepath.Abs(config.UiDir)

	fmt.Printf("config: port %s, ui: %s", config.Port, config.UiDir)

	runServer(config)
}
