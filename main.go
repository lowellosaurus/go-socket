package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Sample project to demonstrate how to work with websockets in Go.
// * Each participant will have a score value
// * Each participant can increase their own score
// * When anyone's score increases, all participants receive an updated state.

type state struct {
	connections map[string]*websocket.Conn
	points      map[string]int
}

type stateOut struct {
	Id     string
	Points []int
}

func (s *state) addParticipant() string {
	id := uuid.New().String()
	s.points[id] = 0

	return id
}

func (s *state) addConnection(id string, conn *websocket.Conn) {
	fmt.Printf("Adding connection for '%s'\n", id)
	s.connections[id] = conn
}

func (s *state) removeParticipant(id string) {
	delete(s.connections, id)
	delete(s.points, id)
}

func (s *state) incrementScore(id string) {
	s.points[id] = s.points[id] + 1
}

func (s *state) serialize(id string) stateOut {
	pts := []int{}
	for _, k := range sortStringIntMapKeys(s.points) {
		pts = append(pts, s.points[k])
	}

	return stateOut{
		Id:     id,
		Points: pts,
	}
}

func (s *state) sendStateToConnections() error {
	fmt.Printf("Sending state update to %d connections\n", len(s.connections))

	for id, conn := range s.connections {
		out := s.serialize(id)
		bytes, err := json.Marshal(out)
		if err != nil {
			return err
		}

		err = conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func sortStringIntMapKeys(m map[string]int) (k []string) {
	for key := range m {
		k = append(k, key)
	}
	sort.Strings(k)

	return k
}

func main() {
	s := &state{
		points:      map[string]int{},
		connections: map[string]*websocket.Conn{},
	}

	// Serve SPA assets.
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/add_participant", addParticipantHandler(s))
	http.HandleFunc("/increment_score", incrementScoreHandler(s))
	http.HandleFunc("/websocket", websocketHandler(s))

	log.Fatal(http.ListenAndServe(":8888", nil))
}

func addParticipantHandler(s *state) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := s.addParticipant()
		out := s.serialize(id)

		bytes, err := json.Marshal(out)
		if err != nil {
			respondWithError(w, fmt.Sprintf("Can't serialize %#v", out))
		}

		io.WriteString(w, string(bytes))
		s.sendStateToConnections()
	}
}

func incrementScoreHandler(s *state) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		s.incrementScore(id)

		out := s.serialize(id)
		bytes, err := json.Marshal(out)
		if err != nil {
			respondWithError(w, fmt.Sprintf("Can't serialize %#v", out))
		}

		io.WriteString(w, string(bytes))
		s.sendStateToConnections()
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(s *state) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Add this new connection to the state.
		id := req.URL.Query().Get("id")
		s.addConnection(id, conn)
		s.sendStateToConnections()

		// TODO: When will we close this connection? Can't close it here
		// because we need it around to send updates to.
		//
		// defer conn.Close()
	}
}

func respondWithError(w http.ResponseWriter, s string) {
	// TODO: We should probably also be responding with a 5xx
	io.WriteString(w, s)
	fmt.Print(s)
}
