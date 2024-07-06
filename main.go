package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PlayerInfo struct {
	Player int
	//Suit   int // 1: rock 2: paper 3: scissor 0: idle
}

type RoomInfo struct {
	PlayerA *PlayerInfo
	PlayerB *PlayerInfo
	Choices []Choice
}

type Choice struct {
	SuitA *int
	SuitB *int
}

var (
	rooms = make(map[string]*RoomInfo)
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
	//roomID := r.URL.Path[len("/rooms/"):]
	// Check if the client count is less than 2
	mu.Lock()
	if _, exist := rooms[roomID]; !exist {
		currentPlayer := &PlayerInfo{Player: 0}
		rooms[roomID] = &RoomInfo{PlayerA: currentPlayer}
		fmt.Printf("Player 1 connected")
		//rooms[roomID] = &RoomInfo{Player: 0, Suit: -1}
	}
	if currentRoom, exist := rooms[roomID]; exist && currentRoom.PlayerA.Player > 0 {
		currentPlayer := &PlayerInfo{Player: 1}
		rooms[roomID].PlayerB = currentPlayer
		fmt.Printf("Player 2 connected")
		//rooms[roomID] = &RoomInfo{Player: 0, Suit: -1}
	}

	// if rooms[roomID].Player >= 2 {
	// 	mu.Unlock()
	// 	http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
	// 	return
	// }

	if currentRoom, exist := rooms[roomID]; exist && currentRoom.PlayerB.Player > 0 {
		mu.Unlock()
		http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
		return
	}

	// if rooms[roomID].PlayerA >= 2 {
	// 	mu.Unlock()
	// 	http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
	// 	return
	// }

	//players[roomID].Player++ //increment disini
	//playerID := players[roomID].Player
	//suit := players[roomID].Suit

	mu.Unlock()

	// // Decrement the client count when the function returns
	// defer func() {
	// 	mu.Lock()
	// 	players[roomID].Player--
	// 	fmt.Printf("Client disconnected. Total clients: %d\n", players[roomID].Player)
	// 	mu.Unlock()
	// }()

	// notify := r.Context().Done()
	// go func() {
	// 	<-notify
	// 	mu.Lock()
	// 	players[roomID].Player--
	// 	fmt.Printf("Client disconnected. Total clients: %d\n", players[roomID].Player)
	// 	mu.Unlock()
	// }()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Define a channel to send messages.
	messages := make(chan string)

	// Goroutine to send messages.
	go func() {
		for {
			// Send a message every second.
			time.Sleep(1 * time.Second)
			//messages <- fmt.Sprintf("data: The time is %s\n\n", time.Now().String())
			if len(rooms[roomID].Choices) > 0 {
				lastChoice := rooms[roomID].Choices[len(rooms[roomID].Choices)-1]
				if *lastChoice.SuitA > 0 {
					playerA := rooms[roomID].PlayerA
					messages <- fmt.Sprintf("data: Player %d, Suit %d, Room %s, The time is %s\n\n", playerA, *lastChoice.SuitA, roomID, time.Now().String())
				}
				if *lastChoice.SuitB > 0 {
					playerB := rooms[roomID].PlayerB
					messages <- fmt.Sprintf("data: Player %d, Suit %d, Room %s, The time is %s\n\n", playerB, *lastChoice.SuitA, roomID, time.Now().String())
				}
			}

		}
	}()

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
	if _, exist := rooms[roomID]; !exist {
		http.Error(w, "Invalid room", http.StatusBadRequest)
	}
	if suit > 0 && suit < 4 {
		if rooms[roomID].PlayerA.Player == player {
			rooms[roomID].Choices[len(rooms[roomID].Choices)-1].SuitA = &suit
			fmt.Fprintf(w, "Player A received for room %d\n", suit)
		}
		if rooms[roomID].PlayerB.Player == player {
			rooms[roomID].Choices[len(rooms[roomID].Choices)-1].SuitB = &suit
			fmt.Fprintf(w, "Player B received for room %d\n", suit)
		}
	} else {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
	}

	mu.Unlock()
}
