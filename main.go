package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

type Player struct {
	ID   int
	Suit []int
}

type Room struct {
	ID      string
	Name    string
	Players []Player
}

var (
	roomx = make(map[string]*Room)
	mu    sync.Mutex
)

func main() {
	http.HandleFunc("/rooms/play/", suit)
	http.HandleFunc("/rooms/", events)

	// Serve the index.html file.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Starting server on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func events(w http.ResponseWriter, r *http.Request) {
	// Set the headers related to event streaming.
	pathParts := strings.Split(r.URL.Path[len("/rooms/"):], "/")
	if len(pathParts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	roomID := pathParts[0]
	mu.Lock()

	var currentPlayer *Player
	if _, exist := roomx[roomID]; !exist {
		player := Player{ID: 0, Suit: []int{}}
		currentPlayer = &player
		var room = Room{ID: roomID, Players: []Player{player}}

		roomx[roomID] = &room
		fmt.Printf("Player 1 connected")
	} else {
		currentRoom := roomx[roomID]
		if len(currentRoom.Players) == 2 {
			http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
			return
		} else if len(currentRoom.Players) == 1 {
			player := Player{ID: 1, Suit: []int{}}
			currentPlayer = &player
			roomx[roomID].Players = append(roomx[roomID].Players, player)
			fmt.Printf("Player 2 connected")
		}
	}

	room := roomx[roomID]

	fmt.Printf("isi player %v", room.Players)
	fmt.Printf("player %v", currentPlayer)
	mu.Unlock()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Define a channel to send messages.
	messages := make(chan string)

	// Goroutine to send messages.
	go func(room *Room, player *Player) {
		fmt.Printf("nganu %d", player.ID)
		for {
			// Send a message every second.
			time.Sleep(1 * time.Second)
			messages <- fmt.Sprintf("data: Player %d, Room %s, The time is %s\n\n", player.ID, roomID, time.Now().String())

			// if len(room.Choices) > 0 {
			// 	lastChoice := room.Choices[len(room.Choices)-1]
			// 	if *lastChoice.SuitA > 0 {
			// 		playerA := room.PlayerA
			// 		messages <- fmt.Sprintf("data: Player %d, Suit %d, Room %s, The time is %s\n\n", playerA, *lastChoice.SuitA, roomID, time.Now().String())
			// 	}
			// 	if *lastChoice.SuitB > 0 {
			// 		playerB := room.PlayerB
			// 		messages <- fmt.Sprintf("data: Player %d, Suit %d, Room %s, The time is %s\n\n", playerB, *lastChoice.SuitA, roomID, time.Now().String())
			// 	}
			// }

		}
	}(room, currentPlayer)

	notify := r.Context().Done()
	go func(room *Room, player *Player) {
		fmt.Printf("nganu %d", player.ID)
		<-notify
		mu.Lock()
		// filteredItems, err := Filter(roomx[room.ID].Players, func(item Player) bool {
		// 	return item.ID != player.ID
		// })
		// if err{
		// 	fmt.Printf("Error in disconnected player.")
		// 	return
		// }

		roomx[room.ID].Players = slices.DeleteFunc(roomx[roomID].Players, func(n Player) bool {
			return n.ID != player.ID
		})
		fmt.Printf("Client disconnected. Total clients: %d\n", len(roomx[room.ID].Players))
		mu.Unlock()
	}(room, currentPlayer)

	fmt.Printf("140?")
	// Write messages to the response.
	for msg := range messages {
		fmt.Fprintf(w, msg)
		// Flush the response.
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			log.Println("Unable to flush")
			return
		}
	}
}

func suit(w http.ResponseWriter, r *http.Request) {

	pathParts := strings.Split(r.URL.Path[len("/rooms/a/"):], "/")
	if len(pathParts) != 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	roomID := pathParts[0]
	var player int
	if _, err := fmt.Sscan(pathParts[1], &player); err != nil {
		http.Error(w, "Invalid player", http.StatusBadRequest)
		return
	}
	//player := pathParts[1].
	var suit int
	if _, err := fmt.Sscan(pathParts[2], &suit); err != nil {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
		return
	}

	mu.Lock()
	if _, exist := roomx[roomID]; !exist {
		http.Error(w, "Invalid room", http.StatusBadRequest)
	}
	if suit > 0 && suit < 4 {
		// if roomx[roomID].PlayerA.Player == player {
		// 	roomx[roomID].Choices[len(roomx[roomID].Choices)-1].SuitA = &suit
		// 	fmt.Fprintf(w, "Player A received for room %d\n", suit)
		// }
		// if rooms[roomID].PlayerB.Player == player {
		// 	rooms[roomID].Choices[len(rooms[roomID].Choices)-1].SuitB = &suit
		// 	fmt.Fprintf(w, "Player B received for room %d\n", suit)
		// }
	} else {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
	}

	mu.Unlock()
}
