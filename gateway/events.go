package gateway

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lckrugel/billy-the-bot/bot"
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
	Operation OpCode  `json:"op"`
	Data      any     `json:"d"`
	Sequence  *int    `json:"s"`
	Type      *string `json:"t"`
}

func (eventPayload GatewayEventPayload) String() string {
	return fmt.Sprintf("{ \"op\": %v, \"d\": %v, \"s\": %v, \"t\": %v }",
		eventPayload.Operation.String(),
		eventPayload.Data,
		eventPayload.Sequence,
		eventPayload.Type)
}

func (eventPayload GatewayEventPayload) GetPayloadData() (map[string]any, error) {
	payloadData, ok := eventPayload.Data.(map[string]any)
	if !ok {
		return nil, errors.New("failed to convert payload data to map")
	}
	return payloadData, nil
}

func unmarshalPayload(msg []byte) (GatewayEventPayload, error) {
	var payload GatewayEventPayload
	err := json.Unmarshal(msg, &payload)
	return payload, err
}

func CreateHeartbeatPayload(sequence *int) ([]byte, error) {
	seq := any(sequence)
	heartbeatPayload := GatewayEventPayload{
		Operation: Heartbeat,
		Data:      &seq,
	}
	heartbeatPayloadJSON, err := json.Marshal(heartbeatPayload)
	if err != nil {
		return nil, err
	}
	return heartbeatPayloadJSON, nil
}

func CreateIdentifyPayload(bot bot.Bot) ([]byte, error) {
	identifyData := map[string]any{
		"token": bot.GetSecretKey(),
		"properties": map[string]string{
			"os":      "windows",
			"browser": "billy",
			"device":  "billy",
		},
		"intents": bot.GetIntents(),
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
