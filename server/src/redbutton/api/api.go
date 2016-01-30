package api

type (
	VoterStatus struct {
		Happy bool `json:"happy"`
	}

	LoginResponse struct {
		VoterId string `json:"voterId"`
	}

	RoomStatusChangeEvent struct {
		RoomName        string `json:"name"`
		NumFlags        int    `json:"marks"`
		NumParticipants int    `json:"participants"`
	}

	RoomInfo struct {
		Id        string `json:"id"`
		RoomName  string `json:"name"`
		RoomOwner string `json:"owner"`
	}
)
