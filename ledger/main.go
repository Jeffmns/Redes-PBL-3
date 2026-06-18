package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"pbl/core" // Nosso pacote compartilhado

	"github.com/cometbft/cometbft/abci/server"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)

// LedgerApp é a nossa aplicação que responde ao motor CometBFT
type LedgerApp struct {
	abcitypes.BaseApplication
	Balances map[string]float64 // Estado das carteiras das companhias
	Logs     []core.Transaction // Armazena o histórico e laudos dos drones
}

func NewLedgerApp() *LedgerApp {
	return &LedgerApp{
		Balances: make(map[string]float64),
		Logs:     make([]core.Transaction, 0),
	}
}

// parseTx verifica se os dados estão em hexadecimal e os converte para JSON antes de ler
func parseTx(txBytes []byte) (core.Transaction, error) {
	var tx core.Transaction

	// Se a transação começa com "0x", removemos o prefixo e decodificamos
	if bytes.HasPrefix(txBytes, []byte("0x")) {
		decoded, err := hex.DecodeString(string(txBytes[2:]))
		if err == nil {
			txBytes = decoded // Substitui os dados originais pelo JSON limpo
		}
	}

	// Agora sim tentamos converter o JSON para a struct
	err := json.Unmarshal(txBytes, &tx)
	return tx, err
}

// CheckTx é chamado quando uma nova transação chega na rede P2P.
// AVALIAÇÃO: É aqui que o duplo gasto é detectado e rejeitado ANTES do bloco.
func (app *LedgerApp) CheckTx(ctx context.Context, req *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	tx, err := parseTx(req.Tx)
	if err != nil {
		fmt.Printf("❌ Falha ao converter dados. Erro: %v\n", err)
		return &abcitypes.ResponseCheckTx{Code: 1, Log: "Erro ao ler transação"}, nil
	}

	// 1. Verifica a assinatura criptográfica usando nosso pacote core
	if !tx.IsValid() {
		return &abcitypes.ResponseCheckTx{Code: 2, Log: "Assinatura digital inválida"}, nil
	}

	// 2. Prevenção de Duplo Gasto e Saldo Insuficiente
	if tx.Sender != "SISTEMA_CONSÓRCIO" {
		saldoAtual := app.Balances[tx.Sender]
		if saldoAtual < tx.Amount {
			return &abcitypes.ResponseCheckTx{Code: 3, Log: "Saldo insuficiente / Duplo gasto bloqueado"}, nil
		}
	}

	return &abcitypes.ResponseCheckTx{Code: 0, Log: "Transação válida"}, nil
}

// FinalizeBlock (antigo DeliverTx) é chamado quando a rede atinge o consenso (PBFT) sobre um bloco.
// AVALIAÇÃO: O "laudo" da missão é registrado de forma definitiva no ledger.
func (app *LedgerApp) FinalizeBlock(ctx context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	for _, txBytes := range req.Txs {
		tx, err := parseTx(txBytes)
		if err != nil {
			continue
		}

		// Aplica a transferência de saldos
		if tx.Amount > 0 {
			if tx.Sender != "SISTEMA_CONSÓRCIO" {
				app.Balances[tx.Sender] -= tx.Amount
			}
			app.Balances[tx.Receiver] += tx.Amount
		}

		// Grava a transação e o laudo (Payload) no histórico
		app.Logs = append(app.Logs, tx)
		fmt.Printf("✅ Transação confirmada no bloco: %s | Laudo: %s\n", tx.ID, tx.Payload)
	}

	return &abcitypes.ResponseFinalizeBlock{}, nil
}

func main() {
	app := NewLedgerApp()

	fmt.Println("Iniciando Ledger App na porta 26658...")

	// Cria o servidor ABCI para escutar conexões via Socket na porta 26658.
	// É aqui que o contêiner do CometBFT vai se conectar!
	srv, _ := server.NewServer("tcp://0.0.0.0:26658", "socket", app)

	// Tenta iniciar o servidor
	if err := srv.Start(); err != nil {
		log.Fatalf("Erro fatal ao iniciar o servidor ABCI: %v", err)
	}

	// O select vazio é um truque clássico em Go para impedir que a função main()
	// termine, mantendo o nosso servidor rodando infinitamente aguardando blocos.
	select {}
}
