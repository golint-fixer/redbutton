package redbutton

import (
	"github.com/gorilla/mux"
	"net/http"
	"redbutton/api"
)

func makeRoutes(s *server) http.Handler {
	m := mux.NewRouter()

	r := api.NewRouteWrapper(m)

	// so far only a stub of login service; returns new voterId each time it's called
	r.Post("/api/login", func(c *api.HttpHandlerContext) {
		c.Result(&api.LoginResponse{
			VoterId: uniqueId(),
		})
	})

	r.Post("/api/room", s.createRoom)
	r.Get("/api/room/{roomId}", s.getRoomInfo)
	r.Post("/api/room/{roomId}", s.updateRoomInfo)
	r.Get("/api/room/{roomId}/voter/{voterId}", s.getVoterStatus)
	r.Post("/api/room/{roomId}/voter/{voterId}", s.handleChangeVoterStatus)
	r.Router.Methods("GET").Path("/api/room/{roomId}/voter/{voterId}/events").HandlerFunc(s.roomEventListenerHandler)
	m.PathPrefix("/").Handler(http.FileServer(http.Dir(s.UiDir)))

	return m
}
