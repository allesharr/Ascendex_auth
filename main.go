package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type API_user struct {
	Host        string
	Port        string
	Group       string
	Apikey      string
	Secret      string
	Name        string
	IsConnected bool
}
type ServerPingMessage struct {
	M  string
	Hp int
}

type AnswerPingMessage struct {
	Op string
}

type BestOrderBook struct {
	Ask Order `json:"ask"` //asks.Price > any bids.Price
	Bid Order `json:"bid"`
}

// Order struct
type Order struct {
	Amount float64 `json:"amount"`
	Price  float64 `json:"price"`
}

type AuthRequest struct {
	Op  string
	Id  uuid.UUID
	T   int64
	Key string
	Sig string
}

type AuthResponse struct {
	M    string
	Id   string
	Code int64
	Err  string
}

type DisconnectMessage struct {
	M      string //disconnect
	Code   int
	Reason string
	Info   string
}

type CancelMessageAllOrders struct {
	Op     string //req
	Action string //cancel-All
	args   CancelMessageAllOrdersArgs
}
type CancelMessageAllOrdersArgs struct {
	time   time.Time
	symbol string //param symbol: optional
}
type CancelMessageOrder struct {
	Op     string //req
	Action string //cancel-All
	args   CancelMessageOrderArgs
}
type CancelMessageOrderArgs struct {
	Time     time.Time
	Coid     uuid.UUID //server returned order id
	OrigCoid uuid.UUID //param symbol: cancel target symbol
	Symbol   string
}

type ConnectedMessage struct {
	Op   string
	Type string
}

var channels = []string{"rder", "trades", "ref-px", "bar", "summary", "depth", "bbo"}

var string_to_connect url.URL

const maxTryes = 5

var GlobalWebSocket *websocket.Conn
var ping_chan = make(chan ServerPingMessage)
var message_chan = make(chan BestOrderBook)
var id = uuid.New()

func (t *API_user) Connection() error {
	// conn, _, err := websocket.DefaultDialer.Dial(string_to_connect.String(),
	// 														http.Header{
	// 															"x-auth-key":

	// })
	conn, _, err := websocket.DefaultDialer.Dial(string_to_connect.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	GlobalWebSocket = conn

	time_UTC := time.Now().UTC().UnixMilli()
	string_to_sign := fmt.Sprint(time_UTC) + "stream"
	auth := AuthRequest{
		Op:  "auth",
		Id:  id,
		T:   time_UTC,
		Key: t.Apikey,
		Sig: t.SignString(string_to_sign),
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	done := make(chan struct{})
	server_mess_chan := make(chan ServerPingMessage)

	go func() {
		for {
			_, message, err := GlobalWebSocket.ReadMessage()
			if err != nil {
				break
			}
			var SM ServerPingMessage
			json.Unmarshal(message, &SM)
			if SM.M == "ping" {
				server_mess_chan <- SM
				continue
			}
			var 

		}
	}()

	go func() {
		select {
		case <-done:
			return

		case <-server_mess_chan:
			ans := AnswerPingMessage{
				op: "pong",
			}
			GlobalWebSocket.WriteJSON(ans)

		}
	}()

	conn.WriteJSON(auth)
	return nil
}

func (t *API_user) GetTiemstamp() int64 {
	return 0
}

func (t *API_user) SignString(string_to_sign string) string {
	secret := t.Secret
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(string_to_sign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (t *API_user) Disconnect() {
	// 	auth := AuthRequest{
	// 		Op:  "auth",
	// 		Id:  id,
	// 		T:   time_UTC,
	// 		Key: t.Apikey,
	// 		Sig: t.SignString(string_to_sign),
	// 	}

	// GlobalWebSocket.WriteJSON()
}

func (t *API_user) SubscribeToChannel(symbol string) error {
	return nil
}

func (t *API_user) ReadMessagesFromChannel(ch chan<- BestOrderBook) {

}

func (t *API_user) WriteMessagesToChannel() {

}

func ReadUserConfiguration() API_user {
	user := API_user{}
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		switch parts[0] {
		case "Open":
			user.Apikey = parts[1]

		case "Secret":
			user.Secret = parts[1]

		case "Group":
			user.Group = parts[1]

		case "Name":
			user.Name = parts[1]

		case "Host":
			user.Host = parts[1]
			if len(parts) > 2 {
				user.Port = parts[2]
			}

		}

	}

	user.IsConnected = false
	if user.Apikey == "" || user.Secret == "" {
		panic("Cannot read Keys")
	}
	// fmt.Println(user.Host)
	if user.Host != "" {
		var host string
		if user.Port != "" {
			host = user.Host + ":" + user.Port
		} else {
			host = user.Host
		}
		path := "/" + user.Group + "/api/pro/v1/stream"
		string_to_connect = url.URL{Scheme: "wss", Host: host, Path: path}
		fmt.Println(string_to_connect.String())
	} else {
		panic("Host is not set")
	}
	return user
}

func main() {
	user := ReadUserConfiguration()
	fmt.Println(user)
	err := user.Connection()
	if err != nil {
		fmt.Println(err)
	}
}
