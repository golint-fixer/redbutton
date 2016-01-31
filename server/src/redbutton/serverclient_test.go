package redbutton

import (
	"github.com/gorilla/websocket"
	"github.com/jmcvetta/napping"
	"github.com/stretchr/testify/require"
	"net/http"
	"redbutton/api"
	"testing"
)

type APIClient struct {
	serviceEndpoint string
	wsEndpoint      string
	t               *testing.T
	lastResponse    *napping.Response
	lastError       error
	session         napping.Session
}

func newAPIClient(t *testing.T) *APIClient {
	result := &APIClient{
		t:               t,
		serviceEndpoint: "http://0.0.0.0:" + testServerConfig.Port + "/api",
		wsEndpoint:      "ws://0.0.0.0:" + testServerConfig.Port + "/api",
	}
	result.session.Header = &http.Header{}
	return result
}

// remember request status
func (c *APIClient) remember(response *napping.Response, err error) {
	c.lastResponse = response
	c.lastError = err
}

func (c *APIClient) assertResponse(expectedHTTPCode int) {
	require.NoError(c.t, c.lastError)
	require.Equal(c.t, expectedHTTPCode, c.lastResponse.Status(), c.lastResponse.RawText())
}

func (c *APIClient) login() api.LoginResponse {
	result := api.LoginResponse{}
	c.remember(c.session.Post(c.serviceEndpoint+"/login", &struct{}{}, &result, nil))
	c.assertResponse(http.StatusOK)
	return result
}

// set current user header
func (c *APIClient) setCurrentUser(voterID string) {
	c.session.Header.Set("voter-id", voterID)
}

func (c *APIClient) getVoterStatus(roomID string, voterID string, expectedCode int) api.VoterStatus {
	s := api.VoterStatus{}
	c.remember(c.session.Get(c.serviceEndpoint+"/room/"+ roomID +"/voter/"+ voterID, nil, &s, nil))
	c.assertResponse(expectedCode)
	return s
}

func (c *APIClient) updateVoterStatus(roomID string, voterID string, update api.VoterStatus) api.VoterStatus {
	result := api.VoterStatus{}
	c.remember(c.session.Post(c.serviceEndpoint+"/room/"+ roomID +"/voter/"+ voterID, &update, &result, nil))
	c.assertResponse(200)
	return result
}

func (c *APIClient) createNewRoom(info api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	c.remember(c.session.Post(c.serviceEndpoint+"/room", &info, &result, nil))
	return result
}

func (c *APIClient) getRoomInfo(roomID string) api.RoomInfo {
	result := api.RoomInfo{}
	c.remember(c.session.Get(c.serviceEndpoint+"/room/"+ roomID, nil, &result, nil))
	c.assertResponse(200)
	return result
}

func (c *APIClient) updateRoomInfo(roomID string, update api.RoomInfo) api.RoomInfo {
	result := api.RoomInfo{}
	c.remember(c.session.Post(c.serviceEndpoint+"/room/"+ roomID, &update, &result, nil))
	return result
}

func (c *APIClient) listenForEvents(roomID string, voterID string) *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial(c.wsEndpoint+"/room/"+ roomID +"/voter/"+ voterID +"/events", nil)
	require.NoError(c.t, err)
	return conn
}
