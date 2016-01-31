package server

import (
	"sync"
)

type Rooms struct {
	sync.RWMutex
	rooms map[string]*Room
}

func newRoomsList() *Rooms {
	return &Rooms{rooms: map[string]*Room{}}
}

func (rooms *Rooms) findRoom(id string) *Room {
	rooms.RLock()
	defer rooms.RUnlock()

	if room, ok := rooms.rooms[id]; ok {
		return room
	}
	return nil
}

func (rooms *Rooms) newRoom() *Room {
	rooms.RWMutex.Lock()
	defer rooms.Unlock()

	for {
		id := uniqueID()[:16]

		if _, ok := rooms.rooms[id]; ok {
			continue
		}

		room := NewRoom()
		rooms.rooms[id] = room
		room.id = id
		return room
	}

}
