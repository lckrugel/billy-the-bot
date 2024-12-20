package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type OpCode int

const (
	Dispatch                  OpCode = iota
	Heartbeat                        = 1
	Identify                         = 2
	Presence_Update                  = 3
	Voice_State_Update               = 4
	Resume                           = 6
	Reconect                         = 7
	Request_Guild_Members            = 8
	Invalid_Session                  = 9
	Hello                            = 10
	Heartbeat_ACK                    = 11
	Request_Soundboard_Sounds        = 31
)

var operationName = map[OpCode]string{
	Dispatch:                  "dispatch",
	Heartbeat:                 "heartbeat",
	Identify:                  "identify",
	Presence_Update:           "presence_update",
	Voice_State_Update:        "voice_state_update",
	Resume:                    "resume",
	Reconect:                  "reconect",
	Request_Guild_Members:     "request_guild_members",
	Invalid_Session:           "invalid_session",
	Hello:                     "hello",
	Heartbeat_ACK:             "heartbeat_ack",
	Request_Soundboard_Sounds: "request_soundboard_sounds",
}

func (op OpCode) String() string {
	return operationName[op]
}

type GatewayEventPayload struct {
	Operation OpCode                 `json:"op"`
	Data      map[string]interface{} `json:"d"`
	Sequence  *int                   `json:"s"`
	Type      *string                `json:"t"`
}

func (eventPayload GatewayEventPayload) String() string {
	return fmt.Sprintf("{ \"op\": %v, \"d\": %v, \"s\": %v, \"t\": %v }",
		eventPayload.Operation.String(),
		eventPayload.Data,
		eventPayload.Sequence,
		eventPayload.Type)
}

func unmarshalPayload(msg []byte) (GatewayEventPayload, error) {
	var payload GatewayEventPayload
	err := json.Unmarshal(msg, &payload)
	return payload, err
}

func createHeartbeatPayload(sequence *int) ([]byte, error) {
	seq := map[string]interface{}{"seq": sequence}
	heartbeatPayload := GatewayEventPayload{
		Operation: Heartbeat,
		Data:      seq,
	}
	heartbeatPayloadJSON, err := json.Marshal(heartbeatPayload)
	if err != nil {
		return nil, err
	}
	return heartbeatPayloadJSON, nil
}

func createIdentifyPayload(token string, intents uint64) ([]byte, error) {
	identifyData := map[string]any{
		"token": token,
		"properties": map[string]string{
			"os":      "windows",
			"browser": "billy",
			"device":  "billy",
		},
		"intents": intents,
	}

	identifyPayload := GatewayEventPayload{
		Operation: Identify,
		Data:      identifyData,
	}
	idenfityPayloadJSON, err := json.Marshal(identifyPayload)
	if err != nil {
		return nil, err
	}
	return idenfityPayloadJSON, nil
}

func createResumePayload(client Client) ([]byte, error) {
	resumeData := map[string]any{
		"token":      client.token,
		"session_id": client.session_id,
		"seq":        client.heartbeat_interval,
	}

	resumePayload := GatewayEventPayload{
		Operation: Resume,
		Data:      resumeData,
	}
	resumePayloadJSON, err := json.Marshal(resumePayload)
	if err != nil {
		return nil, err
	}
	return resumePayloadJSON, nil
}

/* Envia um evento do tipo Heartbeat */
func SendHeartbeat(client Client) (time.Time, error) {
	heartbeatPayload, err := createHeartbeatPayload(client.last_sequence)
	if err != nil {
		return time.Now(), err
	}
	log.Println("[heartbeat] sending heartbeat")
	client.conn.WriteMessage(websocket.TextMessage, heartbeatPayload)
	return time.Now(), nil
}

/* Envia um evento do tipo Identify */
func SendIdentify(client Client) error {
	identifyPayload, err := createIdentifyPayload(client.token, client.intents)
	if err != nil {
		return err
	}
	client.conn.WriteMessage(websocket.TextMessage, identifyPayload)
	return nil
}

func SendResume(client Client) error {
	resumePayload, err := createResumePayload(client)
	if err != nil {
		return nil
	}
	client.conn.WriteMessage(websocket.TextMessage, resumePayload)
	return nil
}

func getReconnectionData(e GatewayEventPayload) (reconnect_url, session_id string, err error) {
	if *e.Type != "READY" {
		return "", "", errors.New("event is not of type 'READY'")
	}

	reconnect_url, ok := e.Data["resume_gateway_url"].(string)
	if !ok {
		return "", "", errors.New("could not find 'resume_gateway_url'")
	}

	session_id, ok = e.Data["session_id"].(string)
	if !ok {
		return "", "", errors.New("could not find 'session_id'")
	}

	return reconnect_url, session_id, nil
}
