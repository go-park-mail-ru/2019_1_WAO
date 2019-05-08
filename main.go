package main

import (
	"fmt"
	"net/http"

	game "./game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// type Message struct {
// 	Type    string          `json:"type"`
// 	Payload json.RawMessage `json:"payload,omitempty"`
// }

// type Move struct {
// 	Direction string  `json:"direction"`
// 	Delay     float64 `json:"delay"`
// }

var GameController *game.Game

// // func UpdateField() {
// // 	newBlocks := game.FieldGenerator(20, 100, 5)
// // 	for _, player := range game.Players {
// // 		if err := Rooms["first"][player].WriteJSON(newBlocks); err != nil {
// // 			fmt.Println("Error JSON array encoding", err)
// // 			break
// // 		}
// //
// // }

// // func InitGame() {
// // 	UpdateField()

// // }

// func ConnectPlayerToRoom(socketPlayer *websocket.Conn, roomName string) {
// 	// var ok bool
// 	// if room, ok = Rooms[roomName]; !ok {
// 	// 	fmt.Println("This room already exists!")
// 	// 	return
// 	// }
// 	if Rooms[roomName] == nil {
// 		Rooms[roomName] = make(connections)
// 	}
// 	newPlayer := &game.Player{
// 		X:  0,
// 		Y:  0,
// 		Vx: 0,
// 		Vy: 0.002,
// 		Id: len(game.Players),
// 	}
// 	Rooms[roomName][newPlayer] = socketPlayer
// 	game.Players = append(game.Players, newPlayer)
// 	fmt.Printf("Player was connected to the room *%s*\n", roomName)
// 	fmt.Printf("Count of Rooms['first'] players: %d\n", len(Rooms["first"]))
// 	// fmt.Printlf("Players %s room:\n", room)
// }

func SocketFunc(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Socket connection error", err)
		return
	}
	// go func() {
	// 	ConnectPlayerToRoom(ws, "first") // Connect this player
	// 	ReadSocket(ws)
	// }()
	// !!!
	player := game.NewPlayer(ws)
	go player.Listen()
	GameController.AddPlayer(player)
}

// func SendStatePlayers() {
// 	JsonPlayers, err := json.Marshal(game.Players)
// 	if err != nil {
// 		fmt.Println("Error encoding players", err)
// 		return
// 	}
// 	msg := Message{
// 		Type:    "players",
// 		Payload: JsonPlayers,
// 	}

// 	// var sendingMessage []byte
// 	sendingMessage, err := json.Marshal(msg)

// 	for _, player := range Rooms["first"] {
// 		if err := player.WriteMessage(websocket.TextMessage, sendingMessage); err != nil {
// 			fmt.Println("Error send players was occured", err)
// 		}
// 	}

// 	for _, plr := range game.Players {
// 		fmt.Printf("***Player***\nid: %d, x: %f, y: %f, vx: %f, vy: %f\n", plr.Id, plr.X, plr.Y, plr.Vx, plr.Vy)
// 	}
// }

// func CloseAndRemove(ws *websocket.Conn) {
// 	index := -1
// 	defer ws.Close()
// 	for i, player := range game.Players { // Search cycle
// 		if Rooms["first"][player] == ws { // Find player's socket
// 			index = i
// 			delete(Rooms["first"], player)
// 			break
// 		}
// 	}
// 	if index == -1 {
// 		fmt.Println("Socket not found!")
// 		return
// 	}
// 	game.Players = append(game.Players[:index], game.Players[index+1:]...) // Delete this player
// 	fmt.Println("The player was disconnected correctly!")
// 	fmt.Printf("Count of Rooms['first'] players: %d", len(Rooms["first"]))
// }

// func ReadSocket(ws *websocket.Conn) {
// 	// fmt.Println("TEST")
// 	defer ws.Close()
// 	for {
// 		_, buf, err := ws.ReadMessage()
// 		fmt.Println(string(buf))
// 		if err != nil {
// 			fmt.Println("Error reading message", err)
// 			CloseAndRemove(ws)
// 			return
// 		}
// 		var msg Message
// 		if err := json.Unmarshal(buf, &msg); err != nil {
// 			fmt.Println("Error message parsing", err)
// 			return
// 		}

// 		switch msg.Type {
// 		case "move":
// 			var moving Move
// 			if err := json.Unmarshal([]byte(msg.Payload), &moving); err != nil {
// 				fmt.Println("Moving error was occured", err)
// 				return
// 			}
// 			fmt.Println("Move was received")
// 			fmt.Printf("Direction: %s, dt: %f\n", moving.Direction, moving.Delay)
// 		case "map":
// 			newBlocks := game.FieldGenerator(100, 100, 10)
// 			buffer, err := json.Marshal(newBlocks)
// 			if err != nil {
// 				fmt.Println("Error encoding new blocks", err)
// 				return
// 			}
// 			JsonNewBlocks := Message{
// 				Type:    "map",
// 				Payload: buffer,
// 			}

// 			BlocksToSend, err := json.Marshal(JsonNewBlocks)
// 			if err != nil {
// 				fmt.Println("Error encoding new blocks", err)
// 				return
// 			}
// 			for _, connection := range Rooms["first"] {
// 				err = connection.WriteMessage(websocket.TextMessage, BlocksToSend)
// 				if err != nil {
// 					fmt.Println("Error send new blocks", err)
// 				}
// 			}

// 		// case "players":
// 		case "lose":
// 			fmt.Println("!Player lose!")
// 			CloseAndRemove(ws)
// 		case "blocks":
// 			var blocks []game.Block
// 			if err := json.Unmarshal([]byte(msg.Payload), &blocks); err != nil {
// 				fmt.Println("Moving error was occured", err)
// 				return
// 			}
// 			fmt.Println("Blocks:")
// 			for index, block := range blocks {
// 				fmt.Printf("*Block %d*\nx: %f, y: %f\n", index, block.X, block.Y)
// 			}

// 		}
// 	}
// }

// func WriteSocket(ws *websocket.Conn) {
// 	player := game.FieldGenerator(100, 100, 10.0)

// 	err := ws.WriteJSON(player)
// 	if err != nil {
// 		fmt.Println("Error was occured", err)
// 	}
// }

func MainFunc(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// tmpl, _ := template.ParseFiles("./templates/index.html")
	// tmpl.Execute(w, "")

	// fmt.Fprintf(w, "ok")
	// game.GameLoop() // Init game process
	http.ServeFile(w, r, "index.html")

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", MainFunc)
	router.HandleFunc("/websocket", SocketFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")
	// Rooms = make(map[string]connections)
	GameController = game.NewGame(1) // New GameController
	go GameController.Run()
	http.ListenAndServe(":8080", nil)
}
