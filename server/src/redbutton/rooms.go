package redbutton

import "sync"

type Rooms struct {
	sync.RWMutex
	rooms map[string]*Room
}

func newRoomsList() *Rooms {
	return &Rooms{rooms: map[string]*Room{}}
}

func (this *Rooms) findRoom(id string) *Room {
	this.RLock()
	defer this.RUnlock()

	if room, ok := this.rooms[id]; ok {
		return room
	}
	return nil
}

func (this *Rooms) newRoom() *Room {
	this.Lock()
	defer this.Unlock()

	for {
		id := uniqueId()[:12]

		if _, ok := this.rooms[id]; ok {
			continue
		}

		room := NewRoom()
		this.rooms[id] = room
		room.id = id
		return room
	}

}
