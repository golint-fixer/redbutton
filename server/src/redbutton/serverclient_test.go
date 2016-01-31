package redbutton

import (
	"github.com/gorilla/websocket"
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"net/http"
	"redbutton/api"
	"testing"
)

type ApiClient struct {
	serviceEndpoint string
	wsEndpoint      string
	t               *testing.T
	lastResponse    *napping.Response
	lastError       error
	session         napping.Session
}

func newApiClient(t *testing.T) *ApiClient {
	result := &ApiClient{
		t:               t,
		serviceEndpoint: "http://0.0.0.0:" + testServerConfig.Port + "/api",
		wsEndpoint:      "ws://0.0.0.0:" + testServerConfig.Port + "/api",
	}
	result.session.Header = &http.Header{}
	return result
}

// remember request status
func (this *ApiClient) remember(response *napping.Response, err error) {
	this.lastResponse = response
	this.lastError = err
}

func (this *ApiClient) assertResponse(expectedHttpCode int) {
	require.NoError(this.t, this.lastError)
	require.Equal(this.t, expectedHttpCode, this.lastResponse.Status(), this.lastResponse.RawText())
}

func (this *ApiClient) login() api.LoginResponse {
	result := api.LoginResponse{}
	this.remember(this.session.Post(this.serviceEndpoint+"/login", &struct{}{}, &result, nil))
	this.assertResponse(http.StatusOK)
	return result
}

// set current user header
func (this *ApiClient) setCurrentUser(voterId string) {
	this.session.Header.Set("voter-id", voterId)
}

func (this *ApiClient) getVoterStatus(roomId string, voterId string, expectedCode int) api.VoterStatus {
	s := api.VoterStatus{}
	this.remember(this.session.Get(this.serviceEndpoint+"/room/"+roomId+"/voter/"+voterId, nil, &s, nil))
	this.assertResponse(expectedCode)
	return s
}

func (this *ApiClient) updateVoterStatus(roomId string, voterId string, update api.VoterStatus) api.VoterStatus {
	result := api.VoterStatus{}
	this.remember(this.session.Post(this.serviceEndpoint+"/room/"+roomId+"/voter/"+voterId, &update, &result, nil))
	this.assertResponse(200)
	return result
}

func (this *ApiClient) createNewRoom(info api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	this.remember(this.session.Post(this.serviceEndpoint+"/room", &info, &result, nil))
	return result
}

func (this *ApiClient) getRoomInfo(roomId string) api.RoomInfo {
	result := api.RoomInfo{}
	this.remember(this.session.Get(this.serviceEndpoint+"/room/"+roomId, nil, &result, nil))
	this.assertResponse(200)
	return result
}

func (this *ApiClient) updateRoomInfo(roomId string, update api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	this.remember(this.session.Post(this.serviceEndpoint+"/room/"+roomId, &update, &result, nil))
	return result
}

func (this *ApiClient) listenForEvents(roomId string, voterId string) *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial(this.wsEndpoint+"/room/"+roomId+"/voter/"+voterId+"/events", nil)
	require.NoError(this.t, err)
	return conn
}
