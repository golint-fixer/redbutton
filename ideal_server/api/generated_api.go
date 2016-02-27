package api

// this all should be generated from spec



type NewRoom struct {

}

type RoomStatus struct {

}


type Service interface {
	CreateNewRoom(room NewRoom) RoomStatus
}