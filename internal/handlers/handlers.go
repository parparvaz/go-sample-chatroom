package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
	"strings"
)

type (
	WsJsonResponse struct {
		Action         string   `json:"action"`
		Message        string   `json:"message"`
		MessageType    string   `json:"message_type"`
		ConnectedUsers []string `json:"connected_users"`
	}

	WsPayload struct {
		Action   string              `json:"action"`
		Username string              `json:"username"`
		Message  string              `json:"message"`
		Conn     WebSocketConnection `json:"-"`
	}

	WebSocketConnection struct {
		*websocket.Conn
	}
)

var (
	views = jet.NewSet(
		jet.NewOSFileSystemLoader("./html"),
		jet.InDevelopmentMode(),
	)
	upgradeConnection = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	wsChannel = make(chan WsPayload)
	clients   = make(map[WebSocketConnection]string)
)

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}

}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = "  "

	var response WsJsonResponse
	response.Message = `<em><small>Connect to server </small></em>`
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenToWs(&conn)
}

func ListenToWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error ", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {

		} else {
			payload.Conn = *conn
			wsChannel <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChannel

		switch e.Action {
		case "username":
			clients[e.Conn] = e.Username

			response.Action = "users-list"
			response.ConnectedUsers = getAllUser()

			Broadcast(response)
			break
		case "left":
			delete(clients, e.Conn)
			response.Action = "users-list"
			response.ConnectedUsers = getAllUser()

			Broadcast(response)
			break
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf(`<strong>%s: </strong> %s`, e.Username, e.Message)

			Broadcast(response)
		}
	}
}

func getAllUser() []string {
	var users []string

	for _, username := range clients {
		if strings.Trim(username, " ") != "" {
			users = append(users, username)
		}
	}
	sort.Strings(users)

	return users
}

func Broadcast(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println(err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		return err
	}

	return nil
}
