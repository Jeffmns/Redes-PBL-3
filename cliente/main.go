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

func main() {
	// Se o usuário não passar nenhum argumento, mostramos como usar
	if len(os.Args) < 2 {
		fmt.Println("Uso esperado do CLI do Consórcio:")
		fmt.Println("  gerar-carteira   - Cria um novo par de chaves para uma companhia")
		fmt.Println("  pagar-escolta    - Envia uma requisição de drone assinada para a rede")
		os.Exit(1)
	}

	// Avalia qual comando foi digitado
	switch os.Args[1] {

	case "gerar-carteira":
		gerarCarteira()

	case "pagar-escolta":
		// Definimos as flags que esse comando aceita
		cmd := flag.NewFlagSet("pagar-escolta", flag.ExitOnError)
		remetente := cmd.String("pub", "", "Chave Pública da Companhia (Remetente)")
		privada := cmd.String("priv", "", "Chave Privada para assinar a transação")
		valor := cmd.Float64("valor", 0, "Quantidade de créditos para pagar a escolta")
		laudo := cmd.String("laudo", "Requisição Inicial", "Informação da missão")

		cmd.Parse(os.Args[2:])

		// Valida se os campos obrigatórios foram preenchidos
		if *remetente == "" || *privada == "" || *valor <= 0 {
			fmt.Println("Erro: Você precisa informar a chave pública, privada e um valor maior que zero.")
			cmd.PrintDefaults()
			os.Exit(1)
		}

		pagarEscolta(*remetente, *privada, *valor, *laudo)

	default:
		fmt.Println("Comando não reconhecido.")
		os.Exit(1)
	}
}

// gerarCarteira cria as credenciais que as companhias usarão no sistema
func gerarCarteira() {
	pubKey, privKey, err := core.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Erro ao gerar chaves: %v", err)
	}

	fmt.Println("=== NOVA CARTEIRA CRIADA ===")
	fmt.Println("Chave Pública (Seu Endereço):", pubKey)
	fmt.Println("Chave Privada (Sua Senha):   ", privKey)
	fmt.Println("GUARDE SUA CHAVE PRIVADA COM SEGURANÇA!")
}

// pagarEscolta monta a transação, assina digitalmente e prepara para enviar
func pagarEscolta(pubKey, privKey string, valor float64, laudo string) {
	// 1. Monta a transação com os dados informados
	tx := core.Transaction{
		Sender:   pubKey,
		Receiver: "SISTEMA_CONSORCIO", // Pagando para o consórcio liberar o drone
		Amount:   valor,
		Payload:  laudo,
	}

	// 2. Assina a transação para provar autenticidade (Evita fraudes!)
	err := tx.Sign(privKey)
	if err != nil {
		log.Fatalf("Erro ao assinar transação: %v", err)
	}

	// 3. Converte a transação pronta para JSON
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("Erro ao converter para JSON: %v", err)
	}

	fmt.Println("✅ Transação assinada e pronta para envio!")
	fmt.Printf("ID: %s\n", tx.ID)

	// 4. Prepara a URL do CometBFT (Codificando o JSON em Hexadecimal com prefixo 0x)
	txHex := hex.EncodeToString(txJSON)

	// ATENÇÃO ao "0x" logo depois da primeira aspa!
	cometURL := fmt.Sprintf("http://cometbft-node-1:26657/broadcast_tx_commit?tx=\"0x%s\"", txHex)

	// 5. Envia a requisição HTTP
	fmt.Println("🚀 Enviando transação para a rede P2P...")
	resp, err := http.Get(cometURL)
	if err != nil {
		log.Fatalf("Erro ao conectar com o nó CometBFT: %v", err)
	}
	defer resp.Body.Close()

	// 6. Lê a resposta da rede
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Resposta do Ledger:")
	fmt.Println(string(body))
}
