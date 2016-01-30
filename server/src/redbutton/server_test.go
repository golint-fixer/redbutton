package redbutton
import (
	"testing"
	"github.com/stretchr/testify/require"
	"time"
	"github.com/jmcvetta/napping"
	"redbutton/api"
	"net/http"
	"github.com/gorilla/websocket"
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
	c := newApiClient(t)
	response := c.login()
	require.NotEqual(t, "", response.VoterId)
}

func TestInvalidRoomId(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	resp, err := napping.Get(c.serviceEndpoint + "/room/whatever/voter/" + loginResponse.VoterId, nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, resp.Status(), 404)
}

func TestLowercaseApi(t *testing.T) {
	result := map[string]string{}
	c := newApiClient(t)
	resp, err := napping.Post(c.serviceEndpoint + "/room", map[string]string{"name":"whatever"}, &result, nil)
	require.NoError(t, err)
	require.Equal(t, resp.Status(), 201)
	require.Equal(t, "whatever", result["name"])
}

func TestGetVoterStatus(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	room := c.createNewRoom(api.RoomInfo{RoomName:"another room", RoomOwner:loginResponse.VoterId})


	{
		// get voter status: happy by default
		s := c.getVoterStatus(room.Id, loginResponse.VoterId)
		require.True(t, s.Happy)
	}

	{
		// set status to unhappy
		r := c.updateVoterStatus(room.Id, loginResponse.VoterId, api.VoterStatus{Happy:false})
		require.False(t, r.Happy)
	}

	{
		// should be unhappy now
		s := c.getVoterStatus(room.Id, loginResponse.VoterId)
		require.False(t, s.Happy)
	}

}

func TestNewRoom(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	result := c.createNewRoom(api.RoomInfo{RoomName:"another room", RoomOwner:loginResponse.VoterId})
	c.assertResponse(201)
	require.NotEqual(t, result.Id, "")
	require.Equal(t, result.RoomName, "another room")
	require.Equal(t, result.RoomOwner, loginResponse.VoterId)

	result2 := c.getRoomInfo(result.Id)
	require.Equal(t, result2.RoomName, "another room")
	require.Equal(t, result2.RoomOwner, loginResponse.VoterId)

	_ = c.createNewRoom(api.RoomInfo{RoomName:"", RoomOwner:loginResponse.VoterId})
	c.assertResponse(http.StatusBadRequest)
}

func TestListenForRoomEvents(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	room := c.createNewRoom(api.RoomInfo{RoomName:"another room", RoomOwner:loginResponse.VoterId})
	c.assertResponse(201)


	conn, _, err := websocket.DefaultDialer.Dial(c.wsEndpoint + "/room/"+room.Id+"/voter/"+loginResponse.VoterId+"/events", nil)
	require.NoError(t, err)
	defer conn.Close()

	roomEvent := api.RoomStatusChangeEvent{}

	// room event should be received after initiation of the connection
	err = conn.ReadJSON(&roomEvent)
	require.NoError(t, err)
	require.Equal(t, room.RoomName, roomEvent.RoomName)
	require.Equal(t, 0, roomEvent.NumFlags)
	require.Equal(t, 1, roomEvent.NumListeners)

	// another room even when someone votes in the room
	c.updateVoterStatus(room.Id, loginResponse.VoterId, api.VoterStatus{Happy:false})
	err = conn.ReadJSON(&roomEvent)
	require.NoError(t, err)
	require.Equal(t, 1, roomEvent.NumFlags)

}

// test if our ID generation uses all hex characters evenly
// this test is silly, but whatever
func TestRandomIds(t *testing.T) {
	counts := map[rune] int{}

	// gather statistics about larger amount of unique IDs
	totalCharCount := 0
	for i:=0;i<100;i++ {
		for _, c := range(uniqueId()) {
			counts[c]++
			totalCharCount ++;
		}
	}

	// each char should have been used approximately 1:16 times
	for _,value := range(counts) {
		require.InEpsilon(t,1.0/16.0,float32(value)/float32(totalCharCount),0.1)
	}
}
