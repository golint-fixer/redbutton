package api

type (
	VoterStatus struct {
		Happy bool `json:"happy"`

		// returns true if voter is current owner of the room
		IsOwner bool `json:"owner"`
	}

	LoginResponse struct {
		VoterID string `json:"voterId"`
	}

	RoomInfo struct {
		ID              string `json:"id"`
		RoomName        string `json:"name"`
		NumParticipants int    `json:"participants"`
		NumFlags        int    `json:"marks"`
	}
)
