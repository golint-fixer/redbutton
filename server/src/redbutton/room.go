package redbutton

import (
	"redbutton/api"
	"sync"
)

type RoomListenerMessageHandler func(msg interface{}) error

// an entity that is interested in room events
// receives notifications via provided message handler
type RoomListener struct {
	messageHandler RoomListenerMessageHandler
}

// notifies this room listener that there's a new event
func (this *RoomListener) newEvent(message interface{}) {
	this.messageHandler(message)
}

type Room struct {
	sync.RWMutex
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

// TODO: http layer ideally should not be in this file at all
func roomAsJson(room *Room) *api.RoomInfo {
	info := api.RoomInfo{}
	info.Id = room.id
	info.RoomName = room.name
	info.RoomOwner = room.owner
	return &info
}

func (this *Room) createListener(handler RoomListenerMessageHandler) *RoomListener {

	listener := &RoomListener{messageHandler: handler}

	this.Lock()
	this.listeners[listener] = true
	this.Unlock()

	this.notifyStatusChanged()
	return listener
}

func (this *Room) unregisterListener(listener *RoomListener) {
	this.Lock()
	delete(this.listeners, listener)
	this.Unlock()

	this.notifyStatusChanged()
}

// builds and sends out a RoomStatusChangeEvent to this room
func (this *Room) notifyStatusChanged() {
	this.RLock()
	msg := api.RoomStatusChangeEvent{
		RoomName:     this.name,
		NumFlags:     len(this.unhappyVotes),
		NumListeners: len(this.listeners),
	}
	this.RUnlock()
	this.broadcastMessage(msg)
}

// broadcasts a message to all room listeners
func (this *Room) broadcastMessage(message interface{}) {
	this.RLock()
	defer this.RUnlock()
	for listener, _ := range this.listeners {
		go listener.newEvent(message)
	}
}

func (this *Room) setHappy(voterId string, happy bool) {
	this.Lock()

	if happy {
		delete(this.unhappyVotes, voterId)
	} else {
		this.unhappyVotes[voterId] = happy
	}
	this.Unlock()

	this.notifyStatusChanged()
}
