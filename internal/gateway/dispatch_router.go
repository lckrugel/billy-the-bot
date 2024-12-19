package gateway

import (
	"fmt"
	"log"
)

func dispatch_router(e GatewayEventPayload) {
	if e.Operation != Dispatch {
		return
	}

	switch *e.Type {
	case "READY":
		fmt.Println("Received 'READY' dispatch event")

	case "RESUMED":
		fmt.Println("Received 'RESUMED' dispatch event")

	default:
		log.Printf("event handler for event of type: %v not implemented", *e.Type)
	}
}
