package api

type (
	VoterStatus struct {
		Happy bool `json:"happy"`

		// returns true if voter is current owner of the room
		IsOwner bool `json:"owner"`
	}

	LoginResponse struct {
		VoterId string `json:"voterId"`
	}

	RoomInfo struct {
		Id              string `json:"id"`
		RoomName        string `json:"name"`
		NumParticipants int    `json:"participants"`
		NumFlags        int    `json:"marks"`
	}
)
