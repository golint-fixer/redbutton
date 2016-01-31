package redbutton

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"redbutton/api"
	"strconv"
	"strings"
)

// generate a random ID
func uniqueId() string {
	h := sha256.New()

	for i := 1; i < 10; i++ {
		h.Write([]byte(strconv.Itoa(rand.Int())))
	}

	result := h.Sum([]byte{})
	return fmt.Sprintf("%x", result)
}

type ServerConfig struct {
	Port  string `envconfig:"PORT" default:"8081"`
	UiDir string `envconfig:"UIDIR" default:"."`
}

type server struct {
	ServerConfig
	rooms             *Rooms
	websocketUpgrader websocket.Upgrader
}

/**
this handler reports room events into provided websocket connection
*/
func (this *server) roomEventListenerHandler(resp http.ResponseWriter, req *http.Request) {
	c := api.NewHandler(req)

	// get voter and room IDs; return on any error
	voterId := this.getVoterIdFromRequest(c)
	if voterId == "" {
		return
	}
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	ws, err := this.websocketUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		c.Error(500, "failed to upgrade to websocket connection: "+err.Error())
		return
	}
	defer ws.Close()

	listener := room.createListener(voterId, func(msg interface{}) error {
		err := websocket.WriteJSON(ws, msg)
		if err != nil {
			log.Println("failed sending json: " + err.Error())
		}
		return err
	})
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

// finds room by ID from 'roomId' url parameter; if not found, sets
// http handler to appropriate HTTP error and returns null
func (this *server) lookupRoomFromRequest(c *api.HttpHandlerContext) *Room {
	roomId := c.PathParam("roomId")
	if roomId == "" {
		c.Error(http.StatusBadRequest, "room ID not found")
		return nil
	}

	room := this.rooms.findRoom(roomId)
	if room == nil {
		c.Error(http.StatusNotFound, "room "+roomId+" was not found")
		return nil
	}

	return room
}

func (this *server) getVoterIdFromRequest(c *api.HttpHandlerContext) VoterId {
	// try path param
	voterId := c.PathParam("voterId")
	if voterId != "" {
		return VoterId(voterId)
	}

	// try header
	voterId = c.Req.Header.Get("voter-id")
	if voterId != "" {
		return VoterId(voterId)
	}

	c.Error(http.StatusBadRequest, "voter ID missing")
	return ""
}

func (this *server) getVoterStatus(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	voterId := this.getVoterIdFromRequest(c)
	if voterId == "" {
		return
	}

	c.Result(room.getVoterStatus(voterId))
}

// voter ID comes from request
func (this *server) handleChangeVoterStatus(c *api.HttpHandlerContext) {
	voterId := this.getVoterIdFromRequest(c)
	if voterId == "" {
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
	info := api.RoomInfo{}
	if !c.ParseRequest(&info) {
		return
	}
	voterId := this.getVoterIdFromRequest(c)
	if voterId == "" {
		return
	}

	info.RoomName = strings.TrimSpace(info.RoomName)
	if info.RoomName == "" {
		c.Error(http.StatusBadRequest, "Room name is missing")
		return
	}

	room := this.rooms.newRoom()
	room.owner = voterId
	room.name = info.RoomName

	c.Status(http.StatusCreated).Result(room.calcRoomInfo())
}

func (this *server) getRoomInfo(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	c.Result(room.calcRoomInfo())
}

func (this *server) updateRoomInfo(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	currentUser := this.getVoterIdFromRequest(c)

	update := api.RoomInfo{
		NumFlags: -1, // unspecified
	}

	if !c.ParseRequest(&update) {
		return
	}

	if update.NumFlags == 0 {
		// request to reset flags to zero
		if currentUser != room.owner {
			c.Error(http.StatusForbidden, "operation allowed for room owners only")
			return
		}

		room.setAllToHappy()
	}

	c.Result(room.calcRoomInfo())
}

func runServer(config ServerConfig) {
	s := server{ServerConfig: config, rooms: newRoomsList()}

	s.websocketUpgrader = websocket.Upgrader{}

	http.Handle("/", makeRoutes(&s))
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		panic(err.Error())
	}
}

func Main() {
	config := ServerConfig{}
	envconfig.Process("redbutton", &config)
	config.UiDir, _ = filepath.Abs(config.UiDir)

	fmt.Printf("config: port %s, ui: %s\n", config.Port, config.UiDir)

	runServer(config)
}
