package api

type (
	VoterStatus struct {
		Happy bool `json:"happy"`
	}

	LoginResponse struct {
		VoterId string `json:"voterId"`
	}

	RoomInfo struct {
		Id              string `json:"id"`
		RoomName        string `json:"name"`
		NumParticipants int    `json:"participants"`
		NumFlags        int    `json:"marks"`
		RoomOwner       string `json:"owner"`
	}
)
