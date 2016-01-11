package redbutton
import (
	"github.com/jmcvetta/napping"
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

func assertResponse(t *testing.T, resp *napping.Response, err error, expectedHttpCode int) {
	require.NoError(t, err)
	require.Equal(t, resp.Status(), expectedHttpCode)
}

func TestLogin(t *testing.T) {
	response := LoginResponse{}
	resp, err := napping.Post(serviceEndpoint() + "/login", &struct{}{}, &response, nil)
	assertResponse(t, resp, err, 200)
}

func TestGetVoterStatus(t *testing.T) {
	loginResponse := LoginResponse{}
	{
		// login
		resp, err := napping.Post(serviceEndpoint() + "/login", &struct{}{}, &loginResponse, nil)
		assertResponse(t, resp, err, 200)
	}

	{
		// get voter status: happy by default
		s := VoterStatus{}
		resp, err := napping.Get(serviceEndpoint() + "/voter/" + loginResponse.VoterId, nil, &s, nil)
		assertResponse(t, resp, err, 200)
		require.True(t, s.Happy)
	}

	{
		// set status to unhappy
		s := VoterStatus{Happy:false}
		r := VoterStatus{}
		resp, err := napping.Post(serviceEndpoint() + "/voter/" + loginResponse.VoterId, &s, &r, nil)
		assertResponse(t, resp, err, 200)
		require.False(t, r.Happy)
	}

	{
		s := VoterStatus{}
		resp, err := napping.Get(serviceEndpoint() + "/voter/" + loginResponse.VoterId, nil, &s, nil)
		assertResponse(t, resp, err, 200)
		// should be unhappy now
		require.False(t, s.Happy)
	}

}