package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"pbl/core"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type StatusDrone struct {
	DroneID string `json:"drone_id"`
	Status  string `json:"status"`
	Laudo   string `json:"laudo"`
}

type AlertaSensor struct {
	Setor       string `json:"setor"`
	Coordenadas string `json:"coordenadas"`
	Gravidade   string `json:"gravidade"`
	Timestamp   string `json:"timestamp"`
}

// Estrutura da resposta da API REST nativa do CometBFT
type CometBlockResponse struct {
	Result struct {
		Block struct {
			Data struct {
				Txs []string `json:"txs"`
			} `json:"data"`
		} `json:"block"`
	} `json:"result"`
}

// =========================================================================
// CAMINHO DE VOLTA (AUDITORIA): Lê o Laudo do Drone e Grava na Blockchain
// =========================================================================
func aoReceberLaudoDoDrone(client mqtt.Client, msg mqtt.Message) {
	var status StatusDrone
	json.Unmarshal(msg.Payload(), &status)

	// Só nos interessa gravar quando ele termina a missão (LIVRE) e tem um laudo
	if status.Status == "LIVRE" && status.Laudo != "" {
		fmt.Printf("\n📝 [Despachante] Recebido Laudo do %s: '%s'\n", status.DroneID, status.Laudo)
		fmt.Println("🔐 [Despachante] A gravar Laudo imutável na Blockchain...")

		// Cria uma transação especial de "Sistema" com custo 0, só para gravar o laudo
		tx := core.Transaction{
			Sender:   "SISTEMA_CONSORCIO", // Usa o nome de sistema para não cobrar saldo
			Receiver: status.DroneID,
			Amount:   0, // Não há transferência de dinheiro, apenas registo
			Payload:  status.Laudo,
		}

		// Converte e manda para o CometBFT (igual o cliente do seu colega faz)
		txJSON, _ := json.Marshal(tx)
		txHex := hex.EncodeToString(txJSON)

		cometURL := os.Getenv("COMETBFT_URL")
		if cometURL == "" {
			cometURL = "http://localhost:26657"
		}

		url := fmt.Sprintf("%s/broadcast_tx_sync?tx=\"0x%s\"", cometURL, txHex)
		http.Get(url) // Dispara a gravação na blockchain!
		fmt.Println("✅ [Despachante] Laudo auditado e gravado com sucesso no CometBFT!")
	}
}

func main() {
	fmt.Println("[Despachante] A iniciar Ponte entre Blockchain e MQTT...")
	time.Sleep(5 * time.Second)

	// =========================================================================
	// 1. CONEXÃO MQTT (Para falar com os Drones e o Raft)
	// =========================================================================
	opts := mqtt.NewClientOptions()
	brokersEnv := os.Getenv("BROKERS")
	if brokersEnv == "" {
		brokersEnv = "tcp://localhost:1883"
	}
	listaBrokers := strings.Split(brokersEnv, ",")
	for _, brokerURL := range listaBrokers {
		opts.AddBroker(brokerURL)
	}
	opts.SetClientID("Despachante_Blockchain")

	mqttClient := mqtt.NewClient(opts)
	for {
		token := mqttClient.Connect()
		token.Wait()
		if token.Error() == nil {
			break
		}
		fmt.Println("⚠️ [Despachante] A aguardar Broker MQTT...")
		time.Sleep(3 * time.Second)
	}
	fmt.Println("✅ [Despachante] Conectado ao MQTT!")

	// -> É AQUI QUE COLOCAMOS O CÓDIGO PARA OUVIR O LAUDO! <-
	// Subscreve ao tópico dos drones para apanhar os laudos quando terminam
	mqttClient.Subscribe("drones/status/+", 1, aoReceberLaudoDoDrone)
	fmt.Println("📡 [Despachante] A ouvir os laudos dos drones para auditoria...")

	// =========================================================================
	// 2. CAMINHO DE IDA (ECONOMIA): Lê a Blockchain e despacha os Drones
	// =========================================================================
	cometURL := os.Getenv("COMETBFT_URL")
	if cometURL == "" {
		cometURL = "http://localhost:26657"
	}

	currentHeight := int64(1) // Começa a ler a partir do Bloco 1

	fmt.Println("🔍 [Despachante] A monitorizar a Blockchain à procura de pagamentos...")
	for {
		url := fmt.Sprintf("%s/block?height=%d", cometURL, currentHeight)
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			// Bloco ainda não foi gerado, espera
			time.Sleep(2 * time.Second)
			continue
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var blockData CometBlockResponse
		json.Unmarshal(body, &blockData)

		// Varre todas as transações daquele bloco
		for _, b64Tx := range blockData.Result.Block.Data.Txs {
			tx, err := parseTxFromBase64(b64Tx)

			// Se conseguiu decifrar e é um pagamento para o Consórcio
			if err == nil && tx.Receiver == "SISTEMA_CONSORCIO" && tx.Amount > 0 {
				fmt.Printf("\n💎 [Despachante] Pagamento confirmado na Blockchain! ID: %s\n", tx.ID[:8])

				// Traduz o pagamento para um alerta que o Raft entende
				alerta := AlertaSensor{
					Setor:       "ConsorcioESG",
					Coordenadas: tx.Payload,
					Gravidade:   "ALTA", // Se pagou, a prioridade é alta!
					Timestamp:   time.Now().Format(time.RFC3339),
				}

				payloadMQTT, _ := json.Marshal(alerta)
				topico := "setor/ConsorcioESG/emergencia"

				fmt.Printf("🚀 [Despachante] A enviar ordem de despacho para o Raft via MQTT (%s)\n", topico)
				mqttClient.Publish(topico, 1, false, payloadMQTT)
			}
		}

		// Avança para o próximo bloco
		currentHeight++
	}
}

// Função auxiliar para descodificar o que vem do CometBFT
func parseTxFromBase64(b64 string) (core.Transaction, error) {
	var tx core.Transaction
	txBytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return tx, err
	}

	// Tira o prefixo "0x" (mesma lógica do ledger do colega)
	if len(txBytes) > 2 && string(txBytes[:2]) == "0x" {
		decoded, err := hex.DecodeString(string(txBytes[2:]))
		if err == nil {
			txBytes = decoded
		}
	}

	err = json.Unmarshal(txBytes, &tx)
	return tx, err
}