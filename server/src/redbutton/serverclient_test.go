package redbutton

import (
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"testing"
	"redbutton/api"
	"net/http"
)

type ApiClient struct {
	serviceEndpoint string
	t               *testing.T
	lastResponse *napping.Response
	lastError error
}

func newApiClient(t *testing.T) *ApiClient {
	return &ApiClient{
		t:t,
		serviceEndpoint: "http://0.0.0.0:"+testServerConfig.Port+"/api",
	}
}


// remember request status
func (this *ApiClient) remember(response *napping.Response, err error){
	this.lastResponse = response
	this.lastError = err
}

func (this *ApiClient) assertResponse(expectedHttpCode int) {
	require.NoError(this.t, this.lastError)
	require.Equal(this.t, expectedHttpCode, this.lastResponse.Status())
}

func (this *ApiClient) login() api.LoginResponse {
	result := api.LoginResponse{}
	this.remember(napping.Post(this.serviceEndpoint + "/login", &struct{}{}, &result, nil))
	this.assertResponse(http.StatusOK)
	return result
}

func (this *ApiClient) getVoterStatus(roomId string, voterId string) api.VoterStatus {
	s := api.VoterStatus{}
	this.remember(napping.Get(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, nil, &s, nil))
	this.assertResponse(200)
	return s
}

func (this *ApiClient) updateVoterStatus(roomId string, voterId string, update api.VoterStatus) api.VoterStatus {
	result := api.VoterStatus{}
	this.remember(napping.Post(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, &update, &result, nil))
	this.assertResponse(200)
	return result
}

func (this *ApiClient) createNewRoom(info api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	this.remember(napping.Post(this.serviceEndpoint+"/room/", &info, &result, nil))
	return result
}

func (this *ApiClient) getRoomInfo(roomId string) api.RoomInfo {
	result := api.RoomInfo{}
	this.remember(napping.Get(this.serviceEndpoint+"/room/"+roomId, nil, &result, nil))
	this.assertResponse(200)
	return result
}


