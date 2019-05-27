package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := "wss://dialogs.herokuapp.com"
	log.Printf("connecting to %s", url)

	addr, err := net.ResolveUDPAddr("udp", "239.228.217.206:54321")
	if err != nil {
		log.Fatalf("udp error: %v", err)
		return
	}

	udp, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("udp error: %v", err)
		return
	}

	defer udp.Close()

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer ws.Close()

	ticker := time.NewTicker(pingPeriod)

	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Fatalf("read error: %v", err)
				}
				break
			}

			log.Printf("received message: %s", message)

			_, err = udp.Write(message)
			if err != nil {
				log.Fatalf("udp write error: %v", err)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Fatalf("write error: %v", err)
				return
			}

			select {
			case <-done:
			}
			return
		case <-ticker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Fatalf("write error: %v", err)
				return
			}
		}
	}
}
