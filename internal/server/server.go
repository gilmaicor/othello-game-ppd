package server

import (
	"log"
	"net/http"
	"othello/internal/game"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan Message)
var gameInstance = game.NewGame()
var mutex = &sync.Mutex{}

type Message struct {
	Type          string   `json:"type"`
	Board         []string `json:"board"`
	CurrentPlayer string   `json:"currentPlayer"`
	Message       string   `json:"message"`
	Color         string   `json:"color"`
	Move          int      `json:"move"`
	ChatMessage   string   `json:"chatMessage"`
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	mutex.Lock()
	if len(clients) >= 2 {
		log.Println("Mais de dois clientes tentaram se conectar. Conexão rejeitada.")
		ws.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "message": "Apenas dois jogadores são permitidos."}`))
		ws.Close()
		mutex.Unlock()
		return
	}

	color := "black"
	if len(clients) == 1 {
		color = "white"
	}
	clients[ws] = color
	mutex.Unlock()

	log.Printf("Nova conexão WebSocket estabelecida: %s", color)

	ws.WriteJSON(Message{Type: "color", Color: color})

	if len(clients) == 2 {
		gameInstance.InitializeBoard()
		log.Println("Dois jogadores conectados. O jogo pode começar.")
		broadcast <- Message{Type: "start", Message: "Dois jogadores conectados. O jogo pode começar.", Board: gameInstance.Board, CurrentPlayer: gameInstance.CurrentPlayer}
	}

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Erro ao ler mensagem: %v", err)
			delete(clients, ws)
			if len(clients) < 2 {
				log.Println("Um jogador desconectou. Aguardando novo jogador.")
				broadcast <- Message{Type: "error", Message: "Um jogador desconectou. Aguardando novo jogador."}
			}
			break
		}
		if msg.Type == "resign" {
			handleResign(clients[ws])
			return
		}
		if msg.Type == "chat" {
			log.Printf("O jogador %s enviou uma mensagem", getColor(clients[ws]))
			broadcast <- Message{Type: "chat", ChatMessage: msg.ChatMessage, Color: clients[ws]}
			continue
		}
		if gameInstance.HandleMove(msg.Move) {
			log.Printf("Movimento recebido do jogador %s: posição %d", getColor(clients[ws]), msg.Move)
			broadcast <- Message{Type: "update", Board: gameInstance.Board, CurrentPlayer: gameInstance.CurrentPlayer}
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func handleResign(color string) {
	opponent := gameInstance.GetOpponent(color)
	var winnerMessage = getColor(opponent) + " vence por desistência!"
	log.Printf("Jogador %s desistiu. %s", getColor(color), winnerMessage)
	broadcast <- Message{Type: "winner", Message: winnerMessage}
}

func getColor(color string) string {
	if color == "black" {
		return "Preto"
	} else {
		return "Branco"
	}
}
