package main

import (
	"log"
	"net/http"
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
var mutex = &sync.Mutex{}
var board = make([]string, 64)
var currentPlayer = "black"

type Message struct {
	Type          string   `json:"type"`
	Board         []string `json:"board"`
	CurrentPlayer string   `json:"currentPlayer"`
	Message       string   `json:"message"`
	Color         string   `json:"color"`
	Move          int      `json:"move"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
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
		initializeBoard()
		log.Println("Dois jogadores conectados. O jogo pode começar.")
		broadcast <- Message{Type: "start", Message: "Dois jogadores conectados. O jogo pode começar.", Board: board, CurrentPlayer: currentPlayer}
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
			handleResign(color)
			return
		}
		log.Printf("Movimento recebido: %v", msg.Move)
		handleMove(msg.Move)
		broadcast <- Message{Type: "update", Board: board, CurrentPlayer: currentPlayer}
	}
}

func handleMessages() {
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

func initializeBoard() {
	for i := 0; i < len(board); i++ {
		board[i] = ""
	}
	board[27] = "white"
	board[28] = "black"
	board[35] = "black"
	board[36] = "white"
	log.Println("Tabuleiro inicializado.")
}

func handleMove(index int) {
	if isValidMove(index, currentPlayer) {
		log.Printf("Movimento válido: Jogador %s para a posição %d", currentPlayer, index)
		makeMove(index, currentPlayer)
		if hasValidMoves(getOpponent(currentPlayer)) {
			currentPlayer = getOpponent(currentPlayer)
			log.Printf("Turno do jogador: %s", currentPlayer)
		} else if !hasValidMoves(currentPlayer) {
			log.Println("Nenhum jogador pode fazer movimentos válidos. Determinando vencedor.")
			determineWinner()
		}
	} else {
		log.Printf("Movimento inválido: Jogador %s para a posição %d", currentPlayer, index)
	}
}

func handleResign(color string) {
	opponent := getOpponent(color)
	var winnerMessage string
	if opponent == "black" {
		winnerMessage = "Preto vence por desistência!"
	} else {
		winnerMessage = "Branco vence por desistência!"
	}
	log.Printf("Jogador %s desistiu. %s", color, winnerMessage)
	broadcast <- Message{Type: "winner", Message: winnerMessage}
}

func isValidMove(index int, player string) bool {
	if board[index] != "" {
		return false
	}
	directions := []struct {
		x, y int
	}{
		{0, 1}, {1, 1}, {1, 0}, {1, -1},
		{0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
	}
	for _, direction := range directions {
		if capturesInDirection(index, direction, player) {
			return true
		}
	}
	return false
}

func capturesInDirection(index int, direction struct{ x, y int }, player string) bool {
	opponent := getOpponent(player)
	hasOpponentBetween := false
	size := 8
	x := index % size
	y := index / size
	x += direction.x
	y += direction.y
	i := y*size + x
	for x >= 0 && x < size && y >= 0 && y < size {
		if board[i] == opponent {
			hasOpponentBetween = true
		} else if board[i] == player {
			return hasOpponentBetween
		} else {
			return false
		}
		x += direction.x
		y += direction.y
		i = y*size + x
	}
	return false
}

func makeMove(index int, player string) {
	flipCells(index, player)
	board[index] = player
}

func flipCells(index int, player string) {
	directions := []struct {
		x, y int
	}{
		{0, 1}, {1, 1}, {1, 0}, {1, -1},
		{0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
	}
	for _, direction := range directions {
		if capturesInDirection(index, direction, player) {
			flipInDirection(index, direction, player)
		}
	}
}

func flipInDirection(index int, direction struct{ x, y int }, player string) {
	opponent := getOpponent(player)
	size := 8
	x := index % size
	y := index / size
	x += direction.x
	y += direction.y
	i := y*size + x
	for x >= 0 && x < size && y >= 0 && y < size {
		if board[i] == opponent {
			board[i] = player
		} else {
			break
		}
		x += direction.x
		y += direction.y
		i = y*size + x
	}
}

func hasValidMoves(player string) bool {
	for i := 0; i < len(board); i++ {
		if isValidMove(i, player) {
			return true
		}
	}
	return false
}

func getOpponent(player string) string {
	if player == "black" {
		return "white"
	}
	return "black"
}

func determineWinner() {
	blackCount, whiteCount := 0, 0
	for _, piece := range board {
		if piece == "black" {
			blackCount++
		} else if piece == "white" {
			whiteCount++
		}
	}
	var winner string
	if blackCount > whiteCount {
		winner = "Preto vence!"
	} else if whiteCount > blackCount {
		winner = "Branco vence!"
	} else {
		winner = "Empate!"
	}
	log.Printf("Determinação do vencedor: %s (Preto: %d, Branco: %d)", winner, blackCount, whiteCount)
	broadcast <- Message{Type: "winner", Message: winner}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("Servidor iniciado em http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
