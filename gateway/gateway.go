package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lckrugel/discord-bot/bot"
)

/* Conecta o bot ao gateway do Discord e ouve por eventos */
func ConnectToGateway(bot bot.Bot) error {
	// Descobre a URL do websocket
	wssURL, err := getWebsocketURL(bot.GetSecretKey())
	if err != nil {
		errMsg := fmt.Sprint("error getting websocket url: ", err)
		return errors.New(errMsg)
	}

	// Estabelece a conexão com o Gateway
	conn, resp, err := websocket.DefaultDialer.Dial(wssURL, nil)
	if err != nil {
		errMsg := fmt.Sprint("error establishing connection to gateway: ", err)
		return errors.New(errMsg)
	}
	defer conn.Close()

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return errors.New(errMsg)
	}

	gatewayEvents := make(chan GatewayEventPayload, 5)

	go eventListener(conn, gatewayEvents) // Começa a ouvir por eventos

	helloPayload := <-gatewayEvents
	if helloPayload.Operation != Hello {
		return errors.New("didn't receive Hello event")
	}
	payloadData, err := helloPayload.GetPayloadData()
	if err != nil {
		errMsg := fmt.Sprint("error getting Hello payload: ", err)
		return errors.New(errMsg)
	}

	// Envia Identify terminando o 'handshake'
	err = sendIdentify(conn, bot)
	if err != nil {
		errMsg := fmt.Sprint("error sending Identify event: ", err)
		return errors.New(errMsg)
	}

	// Usando o heartbeat interval recebido no hello inicia a troca de heartbeats
	heartbeat_interval := payloadData["heartbeat_interval"].(float64)
	err = handleHeartbeat(conn, heartbeat_interval, gatewayEvents)
	if err != nil {
		return err
	}

	return nil
}

/* Recebe a URL de websockets que deve ser usada na conexão */
func getWebsocketURL(api_key string) (string, error) {
	// Forma o request
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/gateway/bot", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bot "+api_key)

	// Envia o request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	// Lê a resposta
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(string(bodyBytes))
	}

	// Transforma o corpo em um map
	var bodyMap map[string]any
	err = json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		return "", err
	}

	// Busca no corpo a URL
	url, ok := bodyMap["url"].(string)
	if !ok {
		return "", errors.New("invalid url")
	}
	return url, nil
}

/* Escuta por eventos na conexão e os envia no canal */
func eventListener(conn *websocket.Conn, ch chan<- GatewayEventPayload) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		responsePayload, err := unmarshalPayload(msg)
		if err != nil {
			return err
		}

		ch <- responsePayload
	}
}

/* Lida com o envio periódico dos "heartbeats" para manter a conexão */
func handleHeartbeat(conn *websocket.Conn, interval float64, ch <-chan GatewayEventPayload) error {
	// Envia o primeiro heartbeat
	log.Println("start sending heartbeats...")
	jitter := rand.Float64() // Intervalo aleatorio antes de começar a enviar heartbeat
	intervalDuration := time.Duration(time.Millisecond * time.Duration(interval))
	time.Sleep(time.Duration(intervalDuration.Milliseconds() * int64(jitter)))
	var lastSeq *int = nil
	lastHeartbeartSentAt, err := sendHeartbeat(conn, lastSeq)
	if err != nil {
		errMsg := fmt.Sprint("failed to send heartbeat: ", err)
		return errors.New(errMsg)
	}

	// Loop de envio de heartbeats
	for lastEvent := range ch {
		if time.Since(lastHeartbeartSentAt) > intervalDuration {
			return errors.New("didn't receive heartbeat ack")
		}
		lastSeq = lastEvent.Sequence

		switch lastEvent.Operation {
		case Heartbeat_ACK:
			log.Print("received heartbeat ack")
			time.Sleep(intervalDuration)
			lastHeartbeartSentAt, err = sendHeartbeat(conn, lastSeq)
			if err != nil {
				errMsg := fmt.Sprint("failed to send heartbeat: ", err)
				return errors.New(errMsg)
			}

		case Heartbeat:
			log.Print("received heartbeat")
			lastHeartbeartSentAt, err = sendHeartbeat(conn, lastSeq)
			if err != nil {
				errMsg := fmt.Sprint("failed to send heartbeat: ", err)
				return errors.New(errMsg)
			}
		}
	}
	return nil
}

/* Envia um evento do tipo Heartbeat */
func sendHeartbeat(conn *websocket.Conn, seq *int) (time.Time, error) {
	heartbeatPayload, err := CreateHeartbeatPayload(seq)
	if err != nil {
		return time.Now(), err
	}
	log.Println("sending heartbeat")
	conn.WriteMessage(websocket.TextMessage, heartbeatPayload)
	return time.Now(), nil
}

/* Envia um evento do tipo Identify */
func sendIdentify(conn *websocket.Conn, bot bot.Bot) error {
	identifyPayload, err := CreateIdentifyPayload(bot)
	if err != nil {
		return err
	}
	conn.WriteMessage(websocket.TextMessage, identifyPayload)
	return nil
}
