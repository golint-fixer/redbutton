package redbutton
import (
	"redbutton/api"
	"github.com/gorilla/websocket"
	"log"
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

func NewRoom() *Room {
	return &Room{
		listeners:    map[*RoomListener]bool{},
		unhappyVotes: map[string]bool{},
	}
}

func roomAsJson(room *Room) *api.RoomInfo {
	info := api.RoomInfo{}
	info.Id = room.id
	info.RoomName = room.name
	info.RoomOwner = room.owner
	return &info
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
