package redbutton

import (
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"testing"
)

type ApiClient struct {
	serviceEndpoint string
	t               *testing.T
}

func (this *ApiClient) assertResponse(resp *napping.Response, err error, expectedHttpCode int) {
	require.NoError(this.t, err)
	require.Equal(this.t, resp.Status(), expectedHttpCode)
}

func (this *ApiClient) login() LoginResponse {
	result := LoginResponse{}
	resp, err := napping.Post(this.serviceEndpoint + "/login", &struct{}{}, &result, nil)
	this.assertResponse(resp, err, 200)
	return result
}

func (this *ApiClient) getVoterStatus(roomId string, voterId string) VoterStatus {
	s := VoterStatus{}
	resp, err := napping.Get(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, nil, &s, nil)
	this.assertResponse(resp, err, 200)
	return s
}

func (this *ApiClient) updateVoterStatus(roomId string, voterId string, update VoterStatus) VoterStatus {
	result := VoterStatus{}
	resp, err := napping.Post(this.serviceEndpoint + "/room/"+roomId+"/voter/" + voterId, &update, &result, nil)
	this.assertResponse(resp, err, 200)
	return result
}

func newApiClient(t *testing.T) *ApiClient {
	return &ApiClient{
		t:t,
		serviceEndpoint: "http://0.0.0.0:"+testServerConfig.Port,
	}

}


