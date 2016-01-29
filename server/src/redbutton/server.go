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

// generate a random voter ID
func uniqueId() string {
	h := sha256.New()
	result := h.Sum([]byte(strconv.Itoa(rand.Int())))

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
	ws, err := this.websocketUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		c.Error(500, "failed to upgrade to websocket connection: " + err.Error())
		return
	}
	defer ws.Close()

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

	room := this.rooms.findRoom(roomId)
	if room == nil {
		c.Error(http.StatusNotFound, "room " + roomId + " was not found")
		return nil
	}

	return room
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

	room := this.rooms.newRoom(id)
	room.owner = info.RoomOwner
	room.name = info.RoomName

	c.Status(http.StatusCreated).Result(roomAsJson(room))
}

func (this *server) getRoomInfo(c *api.HttpHandlerContext) {
	room := this.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	c.Result(roomAsJson(room))
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

func runServer(config ServerConfig) {
	s := server{ServerConfig: config, rooms: newRoomsList()}

	s.websocketUpgrader = websocket.Upgrader{}

	http.Handle("/", router(&s))
	err := http.ListenAndServe(":" + s.Port, nil)
	if err != nil {
		panic(err.Error())
	}

}

func Main() {
	config := ServerConfig{}
	envconfig.Process("redbutton", &config)
	config.UiDir, _ = filepath.Abs(config.UiDir)

	fmt.Printf("config: port %s, ui: %s", config.Port, config.UiDir)

	runServer(config)
}
