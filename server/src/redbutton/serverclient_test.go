package redbutton

import (
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"testing"
	"redbutton/api"
)

type ApiClient struct {
	serviceEndpoint string
	t               *testing.T
}

func (this *ApiClient) assertResponse(resp *napping.Response, err error, expectedHttpCode int) {
	require.NoError(this.t, err)
	require.Equal(this.t, expectedHttpCode, resp.Status())
}

func (this *ApiClient) login() api.LoginResponse {
	result := api.LoginResponse{}
	resp, err := napping.Post(this.serviceEndpoint + "/login", &struct{}{}, &result, nil)
	this.assertResponse(resp, err, 200)
	return result
}

func (this *ApiClient) getVoterStatus(roomId string, voterId string) api.VoterStatus {
	s := api.VoterStatus{}
	resp, err := napping.Get(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, nil, &s, nil)
	this.assertResponse(resp, err, 200)
	return s
}

func (this *ApiClient) updateVoterStatus(roomId string, voterId string, update api.VoterStatus) api.VoterStatus {
	result := api.VoterStatus{}
	resp, err := napping.Post(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, &update, &result, nil)
	this.assertResponse(resp, err, 200)
	return result
}

func (this *ApiClient) createNewRoom(info api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	resp, err := napping.Post(this.serviceEndpoint+"/room/", &info, &result, nil)
	this.assertResponse(resp, err, 201)
	return result
}

func (this *ApiClient) getRoomInfo(roomId string) api.RoomInfo {
	result := api.RoomInfo{}
	resp, err := napping.Get(this.serviceEndpoint+"/room/"+roomId, nil, &result, nil)
	this.assertResponse(resp, err, 200)
	return result
}

func newApiClient(t *testing.T) *ApiClient {
	return &ApiClient{
		t:t,
		serviceEndpoint: "http://0.0.0.0:"+testServerConfig.Port,
	}
}


