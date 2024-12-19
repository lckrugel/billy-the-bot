package gateway

import (
	"log"

	"github.com/gorilla/websocket"
)

/* Escuta por eventos na conexão e os envia nos canais */
func listener(client *Client) {
	for {
		select {
		case <-client.stopSignal:

		default:
			_, msg, err := client.conn.ReadMessage()
			if err != nil {
				if closeErr, ok := err.(*websocket.CloseError); ok {
					log.Println("Listener: ws connection closed: ", closeErr)
					return
				} else {
					log.Fatalf("error reading gateway message: %v", err)
				}
			}

			msgPayload, err := unmarshalPayload(msg)
			if err != nil {
				log.Fatalf("error parsing gateway message: %v", err)
			}

			if msgPayload.Operation != Dispatch {
				client.connectionEvents <- msgPayload
			} else if *msgPayload.Type == "READY" || *msgPayload.Type == "RESUMED" {
				dispatch_router(msgPayload)
				client.connectionEvents <- msgPayload
			} else {
				dispatch_router(msgPayload)
			}
			client.last_sequence = msgPayload.Sequence
		}
	}
}
