package redbutton
import (
	"testing"
	"github.com/stretchr/testify/require"
	"time"
	"github.com/jmcvetta/napping"
	"redbutton/api"
	"net/http"
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

func Test_invalidRoomId(t *testing.T){
	c := newApiClient(t)
	loginResponse := c.login()
	resp,err := napping.Get(c.serviceEndpoint + "/room/whatever/voter/" + loginResponse.VoterId, nil, nil, nil)
	require.NoError(t,err)
	require.Equal(t,resp.Status(),404)
}

func Test_lowercaseApi(t *testing.T){
	result := map[string]string {}
	c := newApiClient(t)
	resp,err := napping.Post(c.serviceEndpoint + "/room", map[string]string{"name":"whatever"}, &result, nil)
	require.NoError(t,err)
	require.Equal(t,resp.Status(),201)
	require.Equal(t,"whatever",result["name"])
}

func TestGetVoterStatus(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	roomId := "default"

	{
		// get voter status: happy by default
		s := c.getVoterStatus(roomId,loginResponse.VoterId)
		require.True(t, s.Happy)
	}

	{
		// set status to unhappy
		r := c.updateVoterStatus(roomId,loginResponse.VoterId, api.VoterStatus{Happy:false})
		require.False(t, r.Happy)
	}

	{
		// should be unhappy now
		s := c.getVoterStatus(roomId,loginResponse.VoterId)
		require.False(t, s.Happy)
	}

}

func TestNewRoom(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	result := c.createNewRoom(api.RoomInfo{RoomName:"another room",RoomOwner:loginResponse.VoterId})
	c.assertResponse(201)
	require.NotEqual(t,result.Id,"")
	require.Equal(t,result.RoomName,"another room")
	require.Equal(t,result.RoomOwner,loginResponse.VoterId)

	result2 := c.getRoomInfo(result.Id)
	require.Equal(t,result2.RoomName,"another room")
	require.Equal(t,result2.RoomOwner,loginResponse.VoterId)

	_ = c.createNewRoom(api.RoomInfo{RoomName:"",RoomOwner:loginResponse.VoterId})
	c.assertResponse(http.StatusBadRequest)

}