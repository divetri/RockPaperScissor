// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"time"
// )

// // func eventsHandler(w http.ResponseWriter, r *http.Request) {
// // 	//ctx := r.Context()
// // 	w.Header().Set("Access-Control-Allow-Origin", "*")
// // 	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

// // 	w.Header().Set("Contect-Type", "text/event-stream")
// // 	w.Header().Set("Cache-Control", "no-cache")
// // 	w.Header().Set("Connection", "keep-alive")

// // 	notify := r.Context().Done()

// // 	for i := 0; i < 10; i++ {
// // 		select {
// // 		case <-notify:
// // 			return
// // 		default:
// // 			fmt.Fprintf(w, "data: Event %d\n\n", i)
// // 			if f, ok := w.(http.Flusher); ok {
// // 				f.Flush()
// // 			}
// // 			time.Sleep(2 * time.Second)
// // 			//w.(http.Flusher).Flush()
// // 		}

// // 	}

// // 	// closeNotifiy := w.(http.CloseNotifier).CloseNotify()
// // 	// <-closeNotifiy
// // 	//r.Context().Done()
// // }

// func eventsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")

// 	// Define a channel to send messages.
// 	messages := make(chan string)

// 	// Goroutine to send messages.
// 	go func() {
// 		for {
// 			// Send a message every second.
// 			time.Sleep(1 * time.Second)
// 			messages <- fmt.Sprintf("data: The time is %s\n\n", time.Now().String())
// 		}
// 	}()

// 	// Write messages to the response.
// 	for msg := range messages {
// 		fmt.Fprintf(w, msg)
// 		// Flush the response.
// 		if f, ok := w.(http.Flusher); ok {
// 			f.Flush()
// 		} else {
// 			log.Println("Unable to flush")
// 			return
// 		}
// 	}
// }

//	func main() {
//		http.HandlerFunc("/events", eventsHandler)
//		http.ListenAndServe(":3000", nil)
//	}
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RoomInfo struct {
	Player int
	Suit   int // 1: rock 2: paper 3: scissor 0: idle
}

var (
	rooms = make(map[string]*RoomInfo)
	mu    sync.Mutex
)

func main() {
	http.HandleFunc("/rooms/a/", suit)
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
	//playerCount := rooms[roomID]
	mu.Lock()
	if _, exist := rooms[roomID]; !exist {
		rooms[roomID] = &RoomInfo{Player: 0, Suit: -1}
	}
	if rooms[roomID].Player >= 2 {
		mu.Unlock()
		http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
		return
	}
	rooms[roomID].Player++ //increment disini
	playerID := rooms[roomID].Player
	suit := rooms[roomID].Suit

	fmt.Printf("Client connected. Total clients: %d\n", rooms[roomID].Player)
	mu.Unlock()

	// Decrement the client count when the function returns
	defer func() {
		mu.Lock()
		rooms[roomID].Player--
		fmt.Printf("Client disconnected. Total clients: %d\n", rooms[roomID].Player)
		mu.Unlock()
	}()

	notify := r.Context().Done()
	go func() {
		<-notify
		mu.Lock()
		rooms[roomID].Player--
		fmt.Printf("Client disconnected. Total clients: %d\n", rooms[roomID].Player)
		mu.Unlock()
	}()

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
			messages <- fmt.Sprintf("data: Player %d, Suit %d, Room %s, The time is %s\n\n", playerID, suit, roomID, time.Now().String())
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
	if room, exist := rooms[roomID]; !exist || room.Player == 0 {
		http.Error(w, "Invalid room", http.StatusBadRequest)
	}
	if rooms[roomID].Player == player && suit > 0 && suit < 4 {
		rooms[roomID].Suit = suit
		fmt.Fprintf(w, "Input %s received for room %s\n", suit, roomID)
	} else {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
	}

	mu.Unlock()
}
