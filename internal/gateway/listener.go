package gateway

import (
	"log"

	"github.com/gorilla/websocket"
)

/* Escuta por eventos na conexão e os envia nos canais */
func listener(client *Client) {
	log.Println("[listener] starting listener...")
	for {
		select {
		case <-client.stopSignal:
			log.Println("[listener] stopping goroutine...")
			return

		default:
		}

		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				log.Println("[listener] ws connection closed: ", closeErr)
				return // Apenas retorna do listener encerrando a goroutine, a reconexão será tentada pelo CloseHandler
			} else {
				log.Fatalf("[listener] unexpected error reading gateway message: %v", err)
			}
		}

		msgPayload, err := unmarshalPayload(msg)
		if err != nil {
			log.Fatalf("[listener] error parsing gateway message: %v", err)
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
