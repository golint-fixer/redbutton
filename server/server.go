package server

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/viktorasm/redbutton/server/api"

	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

// generate a random ID
func uniqueID() string {
	h := sha256.New()

	for i := 1; i < 10; i++ {
		h.Write([]byte(strconv.Itoa(rand.Int())))
	}

	result := h.Sum([]byte{})
	return fmt.Sprintf("%x", result)
}

type Config struct {
	Port  string `envconfig:"PORT" default:"8081"`
	UIDir string `envconfig:"UIDIR" default:"."`
}

type server struct {
	Config
	rooms             *Rooms
	websocketUpgrader websocket.Upgrader
}

/**
this handler reports room events into provided websocket connection
*/
func (s *server) roomEventListenerHandler(resp http.ResponseWriter, req *http.Request) {
	c := api.NewHandler(req)

	// get voter and room IDs; return on any error
	voterID := s.getVoterIDFromRequest(c)
	if voterID == "" {
		return
	}
	room := s.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	ws, err := s.websocketUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		c.Error(500, "failed to upgrade to websocket connection: "+err.Error())
		return
	}
	defer ws.Close()

	listener := room.createListener(voterID, func(msg interface{}) error {
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
func (s *server) lookupRoomFromRequest(c *api.HTTPHandlerContext) *Room {
	roomID := c.PathParam("roomId")
	if roomID == "" {
		c.Error(http.StatusBadRequest, "room ID not found")
		return nil
	}

	room := s.rooms.findRoom(roomID)
	if room == nil {
		c.Error(http.StatusNotFound, "room "+roomID+" was not found")
		return nil
	}

	return room
}

func (s *server) getVoterIDFromRequest(c *api.HTTPHandlerContext) VoterID {
	// try path param
	voterID := c.PathParam("voterId")
	if voterID != "" {
		return VoterID(voterID)
	}

	// try header
	voterID = c.Req.Header.Get("voter-id")
	if voterID != "" {
		return VoterID(voterID)
	}

	c.Error(http.StatusBadRequest, "voter ID missing")
	return ""
}

func (s *server) getVoterStatus(c *api.HTTPHandlerContext) {
	room := s.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	voterID := s.getVoterIDFromRequest(c)
	if voterID == "" {
		return
	}

	c.Result(room.getVoterStatus(voterID))
}

// voter ID comes from request
func (s *server) handleChangeVoterStatus(c *api.HTTPHandlerContext) {
	voterID := s.getVoterIDFromRequest(c)
	if voterID == "" {
		return
	}
	room := s.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	request := api.VoterStatus{}
	if !c.ParseRequest(&request) {
		return
	}

	room.setHappy(voterID, request.Happy)
	s.getVoterStatus(c)
}

func (s *server) createRoom(c *api.HTTPHandlerContext) {
	info := api.RoomInfo{}
	if !c.ParseRequest(&info) {
		return
	}
	voterID := s.getVoterIDFromRequest(c)
	if voterID == "" {
		return
	}

	info.RoomName = strings.TrimSpace(info.RoomName)
	if info.RoomName == "" {
		c.Error(http.StatusBadRequest, "Room name is missing")
		return
	}

	room := s.rooms.newRoom()
	room.owner = voterID
	room.name = info.RoomName

	c.Status(http.StatusCreated).Result(room.calcRoomInfo())
}

func (s *server) getRoomInfo(c *api.HTTPHandlerContext) {
	room := s.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	c.Result(room.calcRoomInfo())
}

func (s *server) updateRoomInfo(c *api.HTTPHandlerContext) {
	room := s.lookupRoomFromRequest(c)
	if room == nil {
		return
	}

	currentUser := s.getVoterIDFromRequest(c)

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

func runServer(config Config) {
	s := server{Config: config, rooms: newRoomsList()}

	s.websocketUpgrader = websocket.Upgrader{}

	http.Handle("/", makeRoutes(&s))
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		panic(err.Error())
	}
}

func Main() {
	config := Config{}
	envconfig.Process("redbutton", &config)
	config.UIDir, _ = filepath.Abs(config.UIDir)

	fmt.Printf("config: port %s, ui: %s\n", config.Port, config.UIDir)

	runServer(config)
}
