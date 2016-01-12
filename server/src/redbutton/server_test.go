package redbutton
import (
	"testing"
	"github.com/stretchr/testify/require"
	"time"
)


func init() {
	go runServer(ServerConfig{Port:"9001"})
	// wait for server startup
	time.Sleep(time.Duration(100) * time.Millisecond)
}

func serviceEndpoint() string {
	return "http://0.0.0.0:9001"
}


func TestLogin(t *testing.T) {
	c := newApiClient(t)
	response := c.login()
	require.NotEqual(t, "", response.VoterId)
}

func TestGetVoterStatus(t *testing.T) {
	c := newApiClient(t)
	loginResponse := c.login()
	{
		// get voter status: happy by default
		s := c.getVoterStatus(loginResponse.VoterId)
		require.True(t, s.Happy)
	}

	{
		// set status to unhappy
		r := c.updateVoterStatus(loginResponse.VoterId, VoterStatus{Happy:false})
		require.False(t, r.Happy)
	}

	{
		// should be unhappy now
		s := c.getVoterStatus(loginResponse.VoterId)
		require.False(t, s.Happy)
	}

}