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
	"github.com/lckrugel/discord-bot/internal/config"
)

type Client struct {
	conn             *websocket.Conn
	last_sequence    *int
	connectionEvents chan GatewayEventPayload
	// dispatchEvents     chan GatewayEventPayload
	stopSignal         chan struct{}
	token              string
	intents            uint64
	heartbeat_interval int
	session_id         string
	reconnect_url      string
}

func (client *Client) SetReconnectParams(session_id string, url string) {
	client.session_id = session_id
	client.reconnect_url = url
}

func (client *Client) Events() <-chan GatewayEventPayload {
	return client.connectionEvents
}

/* Conecta o bot ao gateway do Discord e ouve por eventos */
func (c *Client) Connect() error {
	// Descobre a URL do websocket
	wssURL, err := getWebsocketURL(c.token)
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
	c.conn = conn

	conn.SetCloseHandler(func(code int, text string) error {
		handleCloseCode(code, text, *c)
		return nil
	})

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		errMsg := fmt.Sprint("failed to switch protocols with status: ", resp.StatusCode)
		return errors.New(errMsg)
	}

	// Cria um canal para enviar um sinal e parar a execução das goroutines
	c.stopSignal = make(chan struct{}, 2)

	c.connectionEvents = make(chan GatewayEventPayload, 1)

	go listener(c) // Começa a ouvir por eventos

	helloPayload := <-c.connectionEvents
	if helloPayload.Operation != Hello {
		return errors.New("didn't receive Hello event")
	}

	// Envia Identify terminando o 'handshake'
	err = SendIdentify(*c)
	if err != nil {
		errMsg := fmt.Sprint("error sending Identify event: ", err)
		return errors.New(errMsg)
	}

	// Recebe o Ready com informações sobre como retomar a conexão
	readyPayload := <-c.connectionEvents
	if readyPayload.Operation != Dispatch || *readyPayload.Type != "READY" {
		return errors.New("didn't receive Ready event")
	}
	c.reconnect_url, c.session_id, err = getReconnectionData(readyPayload)
	if err != nil {
		errMsg := fmt.Sprint("error while geting reconnection data: ", err)
		return errors.New(errMsg)
	}
	log.Printf("Rec_URL: %v | Session_id: %v", c.reconnect_url, c.session_id)

	// Usando o heartbeat interval recebido no hello inicia a troca de heartbeats
	c.heartbeat_interval = int(helloPayload.Data["heartbeat_interval"].(float64))

	go handleHeartbeat(c)

	return nil
}

func (c *Client) Reconnect() {
	log.Println("Attempting to reconnect...")

	// Para a execução das goroutines e fecha a conexão
	c.Disconnect()

	// Reinicia a conexão com o gateway usando a url fornecida
	conn, resp, err := websocket.DefaultDialer.Dial(c.reconnect_url, nil)
	if err != nil {
		log.Print("error restablishing connection to gateway: ", err)
		log.Print("restarting connection...")
		c.Connect()
		return
	}
	c.conn = conn

	conn.SetCloseHandler(func(code int, text string) error {
		handleCloseCode(code, text, *c)
		return nil
	})

	// Espera-se que ocorra a troca de HTTP -> WSS
	if resp.StatusCode != 101 {
		log.Print("failed to switch protocols with status: ", resp.StatusCode)
		log.Print("restarting connection...")
		c.Connect()
		return
	}

	// Recria o canal para os sinais de parada
	c.stopSignal = make(chan struct{}, 2)

	go listener(c) // Recomeça a ouvir os eventos
	SendResume(*c)

	// Checa se foi resumida com sucesso a conexão
	resumedPayload := <-c.connectionEvents
	if *resumedPayload.Type != "RESUMED" {
		log.Print("failed to resume connection")
		log.Print("restarting connection...")
		c.Connect()
		return
	}

	go handleHeartbeat(c)
}

func (c *Client) Disconnect() {
	log.Println("Disconnecting...")

	c.sendStopSignal()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	close(c.connectionEvents)
}

func (client *Client) sendStopSignal() {
	if client.stopSignal != nil {
		close(client.stopSignal)
		client.stopSignal = nil
	}
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		token:   cfg.GetSecretKey(),
		intents: cfg.GetIntents(),
	}
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

/* Lida com o envio periódico dos "heartbeats" para manter a conexão */
func handleHeartbeat(client *Client) error {
	// Envia o primeiro heartbeat
	log.Println("[heartbeat] start sending heartbeats...")
	jitter := rand.Float64() // Intervalo aleatorio antes de começar a enviar heartbeat
	intervalDuration := time.Duration(time.Millisecond * time.Duration(client.heartbeat_interval))

	time.Sleep(time.Duration(intervalDuration.Milliseconds() * int64(jitter)))
	client.last_sequence = nil
	lastHeartbeartSentAt, err := SendHeartbeat(*client)
	if err != nil {
		errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
		return errors.New(errMsg)
	}

	// Loop de envio de heartbeats
	for lastEvent := range client.connectionEvents {
		select {
		case <-client.stopSignal:
			return nil

		default:
			if time.Since(lastHeartbeartSentAt) > intervalDuration {
				client.Reconnect()
			}
			client.last_sequence = lastEvent.Sequence

			switch lastEvent.Operation {
			case Heartbeat_ACK:
				log.Print("[heartbeat] received heartbeat ack")
				time.Sleep(intervalDuration)
				lastHeartbeartSentAt, err = SendHeartbeat(*client)
				if err != nil {
					errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
					return errors.New(errMsg)
				}

			case Heartbeat:
				log.Print("[heartbeat] received heartbeat")
				lastHeartbeartSentAt, err = SendHeartbeat(*client)
				if err != nil {
					errMsg := fmt.Sprint("[heartbeat] failed to send heartbeat: ", err)
					return errors.New(errMsg)
				}
			}
		}
	}
	return nil
}

func handleCloseCode(code int, text string, c Client) {
	log.Printf("webSocket closed with code: %d, reason: %s", code, text)
	// Se for possível, tenta a reconexão, se não cria uma nova conexão
	if code > 4010 {
		c.Connect()
	} else {
		c.Reconnect()
	}
}
