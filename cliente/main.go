package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"pbl/core"
)

// Estrutura para salvar a carteira num arquivo local
type Wallet struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const arquivoCarteira = "carteira.json"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso esperado do CLI do Consórcio:")
		fmt.Println("  gerar-carteira   - Cria uma nova carteira e salva automaticamente")
		fmt.Println("  pagar-escolta    - Envia uma requisição de drone (lê a carteira salva)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "gerar-carteira":
		gerarCarteira()

	case "pagar-escolta":
		cmd := flag.NewFlagSet("pagar-escolta", flag.ExitOnError)
		valor := cmd.Float64("valor", 0, "Quantidade de créditos para pagar a escolta")
		laudo := cmd.String("laudo", "Requisição Inicial", "Informação da missão")

		cmd.Parse(os.Args[2:])

		if *valor <= 0 {
			fmt.Println("❌ Erro: Você precisa informar um valor maior que zero (ex: -valor 50).")
			os.Exit(1)
		}

		// A MÁGICA: Lê as chaves automaticamente do arquivo!
		carteira := carregarCarteira()
		if carteira.PublicKey == "" || carteira.PrivateKey == "" {
			fmt.Println("❌ Erro: Nenhuma carteira encontrada no sistema.")
			fmt.Println("👉 Rode 'cliente-app gerar-carteira' primeiro para criar sua conta!")
			os.Exit(1)
		}

		pagarEscolta(carteira.PublicKey, carteira.PrivateKey, *valor, *laudo)

	default:
		fmt.Println("Comando não reconhecido.")
		os.Exit(1)
	}
}

// gerarCarteira cria e salva as credenciais no arquivo carteira.json
func gerarCarteira() {
	pubKey, privKey, err := core.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Erro ao gerar chaves: %v", err)
	}

	// Salva as chaves no arquivo
	carteira := Wallet{PublicKey: pubKey, PrivateKey: privKey}
	dadosJSON, _ := json.MarshalIndent(carteira, "", "  ")
	ioutil.WriteFile(arquivoCarteira, dadosJSON, 0644)

	fmt.Println("=== 💼 NOVA CARTEIRA CRIADA COM SUCESSO ===")
	fmt.Println("As suas chaves foram salvas automaticamente (carteira.json)!")
	fmt.Println("O seu Endereço Público é:", pubKey)
	fmt.Println("Agora você já pode usar o comando pagar-escolta diretamente.")
}

// carregarCarteira lê o arquivo carteira.json e devolve as chaves prontas a usar
func carregarCarteira() Wallet {
	var w Wallet
	arquivo, err := ioutil.ReadFile(arquivoCarteira)
	if err == nil {
		json.Unmarshal(arquivo, &w)
	}
	return w
}

func pagarEscolta(pubKey, privKey string, valor float64, laudo string) {
	tx := core.Transaction{
		Sender:   pubKey,
		Receiver: "SISTEMA_CONSORCIO",
		Amount:   valor,
		Payload:  laudo,
	}

	err := tx.Sign(privKey)
	if err != nil {
		log.Fatalf("Erro ao assinar transação: %v", err)
	}

	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("Erro ao converter para JSON: %v", err)
	}

	fmt.Println("✅ Transação assinada automaticamente com a sua carteira salva!")
	fmt.Printf("ID: %s\n", tx.ID[:8]+"...")

	txHex := hex.EncodeToString(txJSON)
	cometURL := fmt.Sprintf("http://cometbft-node-1:26657/broadcast_tx_sync?tx=\"0x%s\"", txHex)

	fmt.Println("🚀 Enviando transação para a rede P2P...")
	resp, err := http.Get(cometURL)
	if err != nil {
		log.Fatalf("Erro ao conectar com o nó CometBFT: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Resposta do Ledger:")
	fmt.Println(string(body))
}
