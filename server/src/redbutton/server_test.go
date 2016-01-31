package redbutton

import (
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"net/http"
	"redbutton/api"
	"testing"
	"time"
)

var testServerConfig = ServerConfig{
	Port: "9001",
}

func init() {
	go runServer(testServerConfig)
	// wait for server startup
	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestLogin(t *testing.T) {
	c := newAPIClient(t)
	response := c.login()
	require.NotEqual(t, "", response.VoterID)
}

func TestInvalidRoomId(t *testing.T) {
	c := newAPIClient(t)
	loginResponse := c.login()
	resp, err := napping.Get(c.serviceEndpoint+"/room/whatever/voter/"+loginResponse.VoterID, nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, resp.Status(), 404)
}

func TestGetVoterStatus(t *testing.T) {
	c := newAPIClient(t)
	loginResponse := c.login()
	c.setCurrentUser(loginResponse.VoterID)
	room := c.createNewRoom(api.RoomInfo{RoomName: "another room"})

	{
		// get voter status: happy by default
		s := c.getVoterStatus(room.ID, loginResponse.VoterID, http.StatusOK)
		require.True(t, s.Happy)
	}

	{
		// set status to unhappy
		r := c.updateVoterStatus(room.ID, loginResponse.VoterID, api.VoterStatus{Happy: false})
		require.False(t, r.Happy)
	}

	{
		// should be unhappy now
		s := c.getVoterStatus(room.ID, loginResponse.VoterID, http.StatusOK)
		require.False(t, s.Happy)
	}

}

func TestNewRoom(t *testing.T) {
	c := newAPIClient(t)
	loginResponse := c.login()
	c.setCurrentUser(loginResponse.VoterID)
	result := c.createNewRoom(api.RoomInfo{RoomName: "another room"})
	c.assertResponse(201)
	require.NotEqual(t, result.ID, "")
	require.Equal(t, result.RoomName, "another room")

	result2 := c.getRoomInfo(result.ID)
	require.Equal(t, result2.RoomName, "another room")

	_ = c.createNewRoom(api.RoomInfo{RoomName: ""})
	c.assertResponse(http.StatusBadRequest)
}

func TestListenForRoomEvents(t *testing.T) {
	c := newAPIClient(t)
	loginResponse := c.login()
	c.setCurrentUser(loginResponse.VoterID)
	room := c.createNewRoom(api.RoomInfo{RoomName: "another room"})
	c.assertResponse(201)

	conn := c.listenForEvents(room.ID, loginResponse.VoterID)
	defer conn.Close()

	roomEvent := api.RoomInfo{}

	// room event should be received after initiation of the connection
	err := conn.ReadJSON(&roomEvent)
	require.NoError(t, err)
	require.Equal(t, room.RoomName, roomEvent.RoomName)
	require.Equal(t, 0, roomEvent.NumFlags)
	require.Equal(t, 1, roomEvent.NumParticipants)

	// another room even when someone votes in the room
	c.updateVoterStatus(room.ID, loginResponse.VoterID, api.VoterStatus{Happy: false})
	err = conn.ReadJSON(&roomEvent)
	require.NoError(t, err)
	require.Equal(t, 1, roomEvent.NumFlags)

}

// test if our ID generation uses all hex characters evenly;
// this test is silly, but whatever
func TestRandomIds(t *testing.T) {
	counts := map[rune]int{}

	// gather statistics about larger amount of unique IDs
	// count amount of each char separately, as well as grand total
	totalCharCount := 0
	for i := 0; i < 100; i++ {
		for _, c := range uniqueID() {
			counts[c]++
			totalCharCount++
		}
	}

	expectedRatio := 1.0 / 16.0
	// each char should have been used approximately 1:16 times
	for _, value := range counts {
		require.InEpsilon(t, expectedRatio, float32(value)/float32(totalCharCount), 0.1)
	}
}

func TestIsOwner(t *testing.T) {
	c := newAPIClient(t)
	owner := c.login()
	u1 := c.login()
	c.setCurrentUser(owner.VoterID)
	room := c.createNewRoom(api.RoomInfo{RoomName: "room X"})

	status := c.getVoterStatus(room.ID, owner.VoterID, http.StatusOK)
	require.True(t, status.IsOwner)
	status = c.getVoterStatus(room.ID, u1.VoterID, http.StatusOK)
	require.False(t, status.IsOwner)
}

func TestResetVotes(t *testing.T) {
	c := newAPIClient(t)
	owner := c.login()
	u1 := c.login()
	u2 := c.login()
	u3 := c.login()
	c.setCurrentUser(owner.VoterID)
	room := c.createNewRoom(api.RoomInfo{RoomName: "room X"})
	c.assertResponse(201)

	// all participants have to listen for events so that their vote counts to room status
	c.listenForEvents(room.ID, u1.VoterID)
	c.listenForEvents(room.ID, u2.VoterID)
	c.listenForEvents(room.ID, u2.VoterID)

	c.updateVoterStatus(room.ID, u1.VoterID, api.VoterStatus{Happy: false})
	c.updateVoterStatus(room.ID, u2.VoterID, api.VoterStatus{Happy: false})
	c.updateVoterStatus(room.ID, u3.VoterID, api.VoterStatus{Happy: true})

	info := c.getRoomInfo(room.ID)
	require.Equal(t, 2, info.NumFlags)

	// reset our client so we don't have current user set at all
	c = newAPIClient(t)

	// this request requires current user header, and it has to be a room owner
	info = c.updateRoomInfo(room.ID, api.RoomInfo{NumFlags: 0})
	c.assertResponse(http.StatusForbidden)

	c.setCurrentUser(u1.VoterID)
	info = c.updateRoomInfo(room.ID, api.RoomInfo{NumFlags: 0})
	c.assertResponse(http.StatusForbidden)

	c.setCurrentUser(owner.VoterID)
	info = c.updateRoomInfo(room.ID, api.RoomInfo{NumFlags: 0})
	c.assertResponse(http.StatusOK)

	require.Equal(t, 0, info.NumFlags)
	info = c.getRoomInfo(room.ID)
	require.Equal(t, 0, info.NumFlags)
}
