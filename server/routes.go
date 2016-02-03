package server

import (
	"net/http"

	"github.com/viktorasm/redbutton/server/api"

	"github.com/gorilla/mux"
)

func makeRoutes(s *server) http.Handler {
	m := mux.NewRouter()

	r := api.NewRouteWrapper(m)

	// so far only a stub of login service; returns new voterId each time it's called
	r.Post("/api/login", func(c *api.HTTPHandlerContext) {
		c.Result(&api.LoginResponse{
			VoterID: uniqueID(),
		})
	})

	r.Post("/api/room", s.createRoom)
	r.Get("/api/room/{roomId}", s.getRoomInfo)
	r.Post("/api/room/{roomId}", s.updateRoomInfo)
	r.Get("/api/room/{roomId}/voter/{voterId}", s.getVoterStatus)
	r.Post("/api/room/{roomId}/voter/{voterId}", s.handleChangeVoterStatus)
	r.Router.Methods("GET").Path("/api/room/{roomId}/voter/{voterId}/events").HandlerFunc(s.roomEventListenerHandler)
	m.PathPrefix("/swagger.spec.yml").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// serve our spec locally so it can be viewable from swagger-ui
		http.ServeFile(w, r, "api/swagger.spec.yml")
	})
	m.PathPrefix("/").Handler(http.FileServer(http.Dir(s.UIDir)))

	return m
}
