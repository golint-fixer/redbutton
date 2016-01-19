package api


type (
	VoterStatus struct {
		Happy bool `json:"happy"`
	}

	LoginResponse struct {
		VoterId string `json:"voterId"`
	}

	RoomStatusChangeEvent struct {
		RoomName     string `json:"name"`
		NumFlags     int    `json:"marks"`
		NumListeners int    `json:"listeners"`
	}

	RoomInfo struct {
		Id        string `json:"id"`
		RoomName  string `json:"name"`
		RoomOwner string `json:"owner"`
	}
)


