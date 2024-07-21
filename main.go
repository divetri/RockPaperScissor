package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SuitResult int

const (
	Draw SuitResult = iota
	Win
	Lose
)

func (s SuitResult) String() string {
	return [...]string{"Draw", "Win", "Lose"}[s]
}

func (s SuitResult) Int() int {
	return int(s)
}

func (s SuitResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Suit int

const (
	Unknown Suit = iota // EnumIndex = 0
	Rock                // EnumIndex = 1
	Paper               // EnumIndex = 2
	Scissor             // EnumIndex = 3
)

func (s Suit) String() string {
	return [...]string{"Unknown", "Rock", "Paper", "Scissor"}[s]
}

func (s Suit) Int() int {
	return int(s)
}

func (s Suit) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (a Suit) Compare(b Suit) SuitResult {
	if a == b {
		return Draw
	} else {
		if a-b == 1 || a-b == -2 {
			return Win
		} else {
			return Lose
		}
	}
}

type Player struct {
	ID   string `json:"id"`
	Suit []Suit `json:"suit"`
}

type RoomState int

const (
	Idle            RoomState = iota // room created, has no player
	WaitingOpponent                  // room created, one player
	Ready                            // room created, two player, no one suited
	WaitingSuit                      // room created, one player suited
	ShowingResult                    // two player suited, waiting for new game
	Rematch                          // go to idle but still saving result
)

func (s RoomState) String() string {
	return [...]string{"Idle", "WaitingOpponent", "Ready", "WaitingSuit", "ShowingResult", "Rematch", "WaitingRematch"}[s]
}

func (s RoomState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type RoomAction int

const (
	DoNothing RoomAction = iota // do nothing
	Join                        // one player join
	Quit                        // one player quit
	Suiting                     // one player suiting
	Rejoin                      // rematch join
)

func (r *Room) AddPlayer() Player {
	if len(r.Players) < 2 {
		p := Player{ID: uuid.NewString(), Suit: []Suit{}}
		r.Players = append(r.Players, p)
		r.Transition(Join)
		return r.Players[len(r.Players)-1]
	}
	return r.Players[len(r.Players)-1]
}

func (r *Room) RejoinPlayerByID(ID string) {
	if r.State == ShowingResult {
		playerA, playerB := r.Players[0], r.Players[1]
		if playerA.ID == ID {
			r.WaitingSuit = &playerB
		} else {
			r.WaitingSuit = &playerA
		}
		r.Transition(Rejoin)
	}
	if r.State == Rematch {
		if r.WaitingSuit != nil && r.WaitingSuit.ID == ID {
			r.Transition(Rejoin)
			r.WaitingSuit = nil
		}
	}
}

func (r *Room) GetPlayerIndexByID(ID string) int {
	return slices.IndexFunc(r.Players, func(n Player) bool {
		return n.ID == ID
	})
}

func (r *Room) GetPlayerByID(ID string) *Player {
	playerIndex := r.GetPlayerIndexByID(ID)
	if playerIndex == -1 {
		return nil
	}
	return &r.Players[playerIndex]
}

func (r *Room) QuitPlayerByID(ID string) {
	r.Players = slices.DeleteFunc(r.Players, func(n Player) bool {
		return n.ID == ID
	})
	if len(r.Players) == 1 {
		r.Players[0].Suit = []Suit{}
		r.Results = []RoomResult{}
		r.WaitingSuit = nil
	}
	r.Transition(Quit)
}

func (r *Room) AddPlayerSuitByID(ID string, suit Suit) {
	if r.State == Ready || r.State == WaitingSuit {
		playerIndex := r.GetPlayerIndexByID(ID)
		if playerIndex != -1 {
			playerA, playerB := r.Players[0], r.Players[1]
			r.Players[playerIndex].Suit = append(r.Players[playerIndex].Suit, suit)
			r.Transition(Suiting)

			if r.State == WaitingSuit {
				if playerA.ID == r.Players[playerIndex].ID {
					r.WaitingSuit = &playerB
				} else {
					r.WaitingSuit = &playerA
				}
			}

			if r.State == ShowingResult {
				r.WaitingSuit = nil
				playerA, playerB := r.Players[0], r.Players[1]
				lastSuitA, lastSuitB := playerA.Suit[len(playerA.Suit)-1], playerB.Suit[len(playerB.Suit)-1]
				suitResult := lastSuitA.Compare(lastSuitB)
				if suitResult == Win {
					r.Results = append(r.Results, RoomResult{
						ID: &playerA.ID,
					})
				}
				if suitResult == Lose {
					r.Results = append(r.Results, RoomResult{

						ID: &playerB.ID,
					})
				}
				if suitResult == Draw {
					r.Results = append(r.Results, RoomResult{
						ID: nil,
					})
				}
			}
		}
	}

}

func (r *Room) Transition(a RoomAction) *Room {
	s := r.State
	if s == Idle && a == Join {
		r.State = WaitingOpponent
		return r
	}
	if s == WaitingOpponent && a == Quit {
		r.State = Idle
		return r
	}
	if s == WaitingOpponent && a == Join {
		r.State = Ready
		return r
	}
	if s == Ready && a == Quit {
		r.State = WaitingOpponent
		return r
	}
	if s == Ready && a == Suiting {
		r.State = WaitingSuit
		return r
	}
	if s == WaitingSuit && a == Quit {
		r.State = WaitingOpponent
		return r
	}
	if s == WaitingSuit && a == Suiting && r.WaitingSuit != nil {
		r.State = ShowingResult
		return r
	}
	if s == ShowingResult && a == Quit {
		r.State = WaitingOpponent //reset result
		return r
	}
	if s == ShowingResult && a == Rejoin {
		r.State = Rematch //player rejoin 1
		return r
	}
	if s == Rematch && a == Rejoin { //player join udah 2
		r.State = Ready
		return r
	}
	if s == Rematch && a == Quit {
		r.State = WaitingOpponent //reset result
		return r
	}
	return r
}

type RoomResult struct {
	ID *string
}

type Room struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Players     []Player     `json:"players"`
	State       RoomState    `json:"state"`
	Results     []RoomResult `json:"results"`
	WaitingSuit *Player      `json:"waitingSuit"`
}

type Message struct {
	Player Player `json:"player"`
	Room   Room   `json:"room"`
}

var (
	roomx = make(map[string]*Room)
	mu    sync.Mutex
)

func main() {
	//TODO:
	//http.HandleFunc("/rooms/list/", getrooms) //display rooms on lobby
	//http.HandleFunc("/rooms/join/{roomID}", joinroom) //join room by link from lobby
	//http.HandleFunc("/rooms/create/", createroom) //idleroom with no initiate player
	//while runningtime < 4 minute of idle state == idle -> dihitung berapa menit idle

	http.HandleFunc("/rooms/rematch/", rematch) //rooms/rematch/roomid/playerid
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
	var currentPlayer Player
	// when currentRoom is not exist, so make the room and first player
	if _, exist := roomx[roomID]; !exist {

		var room = Room{
			ID:      roomID,
			Players: []Player{},
			State:   Idle,
			Results: []RoomResult{},
		}

		currentPlayer = room.AddPlayer()
		roomx[roomID] = &room
		fmt.Printf("\nNew Player 1 connected with ID: %s", currentPlayer.ID)
		mu.Unlock()

	} else {
		currentRoom := roomx[roomID]
		if currentRoom.State == Ready || currentRoom.State == WaitingSuit || currentRoom.State == ShowingResult {
			fmt.Printf("\nToo many clients connected: %s", currentPlayer.ID)
			mu.Unlock()
			http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
			return
		}

		if currentRoom.State == WaitingOpponent {
			currentPlayer = roomx[roomID].AddPlayer()
			fmt.Printf("\nNew Player 2 connected with ID: %s", currentPlayer.ID)
			mu.Unlock()
		}

		if currentRoom.State == Idle {
			currentPlayer = roomx[roomID].AddPlayer()
			fmt.Printf("\nNew Player 1 connected with ID: %s", currentPlayer.ID)
			mu.Unlock()
		}
	}

	room := roomx[roomID]

	notify := r.Context().Done()
	go func(roomID string, playerID string) {
		<-notify
		mu.Lock()
		fmt.Printf("\nCurrent Room %v", roomx[roomID])

		roomx[roomID].QuitPlayerByID(playerID)

		fmt.Printf("\nCurrent Room After Deleted %s : %v", playerID, roomx[roomID])

		fmt.Printf("\nClient disconnected. \nTotal clients: %d\n", len(roomx[roomID].Players))
		mu.Unlock()
	}(room.ID, currentPlayer.ID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Define a channel to send messages.
	messages := make(chan string)

	// Goroutine to send messages.
	go func(roomID string, player Player) {
		for {
			time.Sleep(1 * time.Second)

			if currentPlayer := roomx[roomID].GetPlayerByID(player.ID); currentPlayer != nil {
				if jsonRoom, err := json.Marshal(Message{Room: *roomx[roomID], Player: *currentPlayer}); err == nil {
					messages <- fmt.Sprintf("data: %s\n\n", string(jsonRoom))
				} else {
					log.Println("Cannot Unmarshal Line:272")
				}
			}
		}
	}(room.ID, currentPlayer)

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

	pathParts := strings.Split(r.URL.Path[len("/rooms/play/"):], "/")

	if len(pathParts) != 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	roomID := pathParts[0]
	if len(roomx[roomID].Players) != 2 {
		http.Error(w, "Must be 2 player in the room", http.StatusBadRequest)
		return
	}

	var playerID string
	if _, err := fmt.Sscan(pathParts[1], &playerID); err != nil {
		http.Error(w, "Invalid player", http.StatusBadRequest)
		return
	}

	if !slices.ContainsFunc(roomx[roomID].Players, func(p Player) bool {
		return p.ID == playerID
	}) {
		http.Error(w, "This player is not allowed to submit the suit in this Room.", http.StatusBadRequest)
		return
	}

	var suit int
	if _, err := fmt.Sscan(pathParts[2], &suit); err != nil {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
		return
	}

	if _, exist := roomx[roomID]; !exist {
		http.Error(w, "Invalid room", http.StatusBadRequest)
	}

	if suit > 0 && suit < 4 {
		mu.Lock()

		roomx[roomID].AddPlayerSuitByID(playerID, Suit(suit))
		mu.Unlock()
	} else {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
	}

}

func rematch(w http.ResponseWriter, r *http.Request) {

	pathParts := strings.Split(r.URL.Path[len("/rooms/rematch/"):], "/")

	if len(pathParts) != 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	roomID := pathParts[0]

	var playerID string
	if _, err := fmt.Sscan(pathParts[1], &playerID); err != nil {
		http.Error(w, "Invalid player", http.StatusBadRequest)
		return
	}

	if _, exist := roomx[roomID]; !exist {
		http.Error(w, "Invalid room", http.StatusBadRequest)
	}
	mu.Lock() //semua request mutasinya ditahan
	roomx[roomID].RejoinPlayerByID(playerID)
	mu.Unlock() //release mutasi
}
