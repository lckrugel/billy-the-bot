package gateway

import "log"

/* Escuta por eventos na conexão e os envia nos canais */
func listener(client *Client) {
	for {
		select {
		case <-client.stopSignal:

		default:
			_, msg, err := client.conn.ReadMessage()
			if err != nil {
				log.Fatalf("error reading gateway message: %v", err)
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
