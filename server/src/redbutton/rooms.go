package redbutton

type Rooms struct {
	rooms map[string]*Room
}

func newRoomsList() *Rooms {
	return &Rooms{rooms: map[string]*Room{}}
}

func (this *Rooms) findRoom(id string) *Room {
	if room, ok := this.rooms[id]; ok {
		return room
	}
	return nil
}

func (this *Rooms) newRoom(id string) *Room {
	room := NewRoom()
	this.rooms[id] = room
	room.id = id
	return room
}
