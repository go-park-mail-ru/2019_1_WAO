package game

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

func TestNewGame(t *testing.T) {
	game := NewGame(2)
	if game == nil {
		t.Fatalf("Unexpected nil game\n")
	}
}

func TestRemoveRoom(t *testing.T) {
	game := &Game{
		rooms: []*Room{
			&Room{},
			&Room{},
			&Room{},
			&Room{},
		},
	}
	for _, room := range game.rooms {
		room.game = game
	}

	for i := 0; i < 4; i++ {
		game.RemoveRoom(game.rooms[0])
		if len(game.rooms) != 3-i {
			t.Errorf("Expected count of rooms: %d, but got: %d\n", 3-i, len(game.rooms))
		}
	}
	rooms := []*Room{
		&Room{},
		NewRoom(2, &Game{}),
		NewRoom(3, &Game{}),
		NewRoom(4, &Game{}),
		NewRoom(5, &Game{}),
		NewRoom(6, &Game{}),
	}
	for _, room := range rooms {
		err := game.RemoveRoom(room) // We're trying to remove a room that does not exist
		if err == nil {
			t.Fatalf("Expected error with removing a room, but got nothing")
		}
	}
}
func TestAddRoom(t *testing.T) {
	game := NewGame(4)
	for i := 0; i < 4; i++ {
		game.AddRoom(NewRoom(2, game))
		if len(game.rooms) != i+1 {
			t.Errorf("Expected count of rooms: %d, but got: %d\n", i+1, len(game.rooms))
		}

	}
	for _, room := range game.rooms {
		game.RemoveRoom(room)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func runServer(GameController *Game) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Socket connection error", err)
			return
		}

		// !!!
		player := NewPlayer(ws)
		GameController.AddPlayer(player)
	}))
}

func gameActivate(s *httptest.Server, GameController *Game, done <-chan struct{}) {
	// Convert http://127.0.0.1 to ws://127.0.0.1
	// u := "ws" + strings.TrimPrefix(s.URL, "http")
	go GameController.Run()
	<-done
	s.Close()
}

// func TestGameRun(t *testing.T) {
// 	game := NewGame(2)
// 	done := make(chan struct{})
// 	s := runServer(game)
// 	go gameActivate(s, game, done)
// 	// players := make([]*Player, 3)
// 	// u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
// 	// u.RequestURI()
// 	h :=
// 	for i := 0; i < 2; i++ {
// 		ws, _, err := websocket.DefaultDialer.Dial("ws://"+s.Listener.Addr().String()+"/ws", nil)
// 		if err != nil {
// 			t.Fatalf("%v", err)
// 		}
// 		defer ws.Close()
// 	}
// 	if len(game.rooms) != 1 {
// 		t.Fatalf("Expected: 1 rooms, but got: %d", len(game.rooms))
// 	}
// 	ws, _, err := websocket.DefaultDialer.Dial("ws://"+s.Listener.Addr().String()+"/ws", nil)
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}
// 	defer ws.Close()
// 	// players[2] = NewPlayer(ws)
// 	if len(game.rooms) != 2 {
// 		t.Fatalf("Expected: 2 rooms, but got: %d", len(game.rooms))
// 	}
// 	done <- struct{}{}
// }

func TestAddPlayerToTheGame(t *testing.T) {
	viper.SetConfigFile("../config/test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	player := NewPlayer(nil)
	player2 := NewPlayer(nil)
	player3 := NewPlayer(nil)
	game := NewGame(2)
	go game.Run()
	game.AddPlayer(player)
	game.AddPlayer(player2)
	game.AddPlayer(player3)
}
