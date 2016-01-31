package redbutton

import (
	"redbutton/api"
	"sync"
)

// RoomListenerMessageHandler - a type
type RoomListenerMessageHandler func(msg interface{}) error

type VoterID string

// RoomListener is an entity that is interested in room events
// receives notifications via provided message handler
type RoomListener struct {
	voterID        VoterID
	messageHandler RoomListenerMessageHandler
}

// notifies this room listener that there's a new event
func (l *RoomListener) newEvent(message interface{}) {
	l.messageHandler(message)
}

type Room struct {
	sync.RWMutex
	id        string
	name      string
	owner     VoterID
	listeners map[*RoomListener]bool
	voters    map[VoterID]bool
}

func NewRoom() *Room {
	return &Room{
		listeners: map[*RoomListener]bool{},
		voters:    map[VoterID]bool{},
	}
}

func (room *Room) createListener(voterID VoterID, handler RoomListenerMessageHandler) *RoomListener {

	listener := &RoomListener{voterID: voterID, messageHandler: handler}

	room.Lock()
	room.listeners[listener] = true
	room.Unlock()

	room.notifyStatusChanged()
	return listener
}

func (room *Room) unregisterListener(listener *RoomListener) {
	room.Lock()
	delete(room.listeners, listener)
	room.Unlock()

	room.notifyStatusChanged()
}

func (room *Room) calcRoomInfo() *api.RoomInfo {
	room.RLock()
	defer room.RUnlock()
	// collect IDs of active voters
	activeVoters := map[VoterID]bool{}
	for listener := range room.listeners {
		activeVoters[listener.voterID] = true
	}

	// count votes of active unhappy voters
	numUnhappy := 0
	for voterID, happy := range room.voters {
		if _, ok := activeVoters[voterID]; ok {
			if !happy {
				numUnhappy++
			}
		}
	}

	return &api.RoomInfo{
		ID:              room.id,
		RoomName:        room.name,
		NumFlags:        numUnhappy,
		NumParticipants: len(activeVoters),
	}
}

// builds and sends out a RoomStatusChangeEvent to this room
func (room *Room) notifyStatusChanged() {
	msg := room.calcRoomInfo()
	room.broadcastMessage(msg)
}

// broadcasts a message to all room listeners
func (room *Room) broadcastMessage(message interface{}) {
	room.RLock()
	defer room.RUnlock()
	for listener := range room.listeners {
		go listener.newEvent(message)
	}
}

func (room *Room) setHappy(voterID VoterID, happy bool) {
	room.Lock()
	room.voters[voterID] = happy
	room.Unlock()

	room.notifyStatusChanged()
}

func (room *Room) setAllToHappy() {
	room.Lock()
	for key := range room.voters {
		room.voters[key] = true
	}
	room.Unlock()

	room.notifyStatusChanged()
}

func (room *Room) getVoterStatus(voterID VoterID) *api.VoterStatus {
	result := api.VoterStatus{}
	result.Happy = true
	if happy, ok := room.voters[voterID]; ok {
		result.Happy = happy
	}
	result.IsOwner = voterID == room.owner
	return &result
}
