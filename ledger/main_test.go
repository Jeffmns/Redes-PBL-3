package main

import (
	"context"
	"encoding/json"
	"testing"

	"pbl/core"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

func TestFinalizeBlockReturnsOneTxResultPerTransaction(t *testing.T) {
	app := NewLedgerApp()

	senderPub, senderPriv, err := core.GenerateKeyPair()
	if err != nil {
		t.Fatalf("gerar chave: %v", err)
	}

	tx := core.Transaction{
		Sender:   senderPub,
		Receiver: "SISTEMA_CONSORCIO",
		Amount:   10,
		Payload:  "teste",
	}
	if err := tx.Sign(senderPriv); err != nil {
		t.Fatalf("assinar transação: %v", err)
	}

	payload, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	resp, err := app.FinalizeBlock(context.Background(), &abcitypes.RequestFinalizeBlock{Txs: [][]byte{payload}})
	if err != nil {
		t.Fatalf("finalizar: %v", err)
	}
	if len(resp.TxResults) != 1 {
		t.Fatalf("esperava 1 resultado de tx, mas recebeu %d", len(resp.TxResults))
	}
}

func TestRepeatedPaymentsAreAcceptedAfterInitialBalanceIsExhausted(t *testing.T) {
	app := NewLedgerApp()

	senderPub, senderPriv, err := core.GenerateKeyPair()
	if err != nil {
		t.Fatalf("gerar chave: %v", err)
	}

	makeTx := func(amount float64) []byte {
		tx := core.Transaction{
			Sender:   senderPub,
			Receiver: "SISTEMA_CONSORCIO",
			Amount:   amount,
			Payload:  "pagamento-escola",
		}
		if err := tx.Sign(senderPriv); err != nil {
			t.Fatalf("assinar transação: %v", err)
		}
		data, err := json.Marshal(tx)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return data
	}

	firstTx := makeTx(1000)
	checkResp, err := app.CheckTx(context.Background(), &abcitypes.RequestCheckTx{Tx: firstTx})
	if err != nil {
		t.Fatalf("checktx 1: %v", err)
	}
	if checkResp.Code != 0 {
		t.Fatalf("checktx 1 inesperado: code=%d log=%s", checkResp.Code, checkResp.Log)
	}

	_, err = app.FinalizeBlock(context.Background(), &abcitypes.RequestFinalizeBlock{Txs: [][]byte{firstTx}})
	if err != nil {
		t.Fatalf("finalizar 1: %v", err)
	}

	secondTx := makeTx(1000)
	checkResp, err = app.CheckTx(context.Background(), &abcitypes.RequestCheckTx{Tx: secondTx})
	if err != nil {
		t.Fatalf("checktx 2: %v", err)
	}
	if checkResp.Code != 0 {
		t.Fatalf("checktx 2 deveria ter sido aceito, mas retornou code=%d log=%s", checkResp.Code, checkResp.Log)
	}
}
