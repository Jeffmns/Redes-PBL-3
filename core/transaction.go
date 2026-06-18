package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Transaction struct {
	ID        string  // Hash único da transação
	Sender    string  // Chave pública de quem envia (Companhia ou Drone)
	Receiver  string  // Chave pública de quem recebe (Consórcio ou vazio em caso de laudo)
	Amount    float64 // Quantidade de "créditos operacionais" transferidos
	Payload   string  // Dados extras: aqui vai o "laudo" da missão (ex: "rota segura", "obstáculo")
	Signature string  // Assinatura criptográfica para provar que o Sender autorizou a operação
	Timestamp int64   // Data e hora da criação
}

// GetSignableString junta os dados fundamentais da transação em um único texto estruturado.
// ATENÇÃO: Deixamos o ID e a Signature de fora, pois eles são gerados DEPOIS com base neste texto.
func (tx *Transaction) GetSignableString() string {
	return fmt.Sprintf("%s->%s:%f[%s]%d", tx.Sender, tx.Receiver, tx.Amount, tx.Payload, tx.Timestamp)
}

// Sign tenta assinar a transação usando a chave privada do remetente
func (tx *Transaction) Sign(privateKeyHex string) error {
	dataToSign := tx.GetSignableString()
	sig, err := SignData(dataToSign, privateKeyHex)
	if err != nil {
		return err
	}
	tx.Signature = sig

	// Aproveitamos para gerar o ID único da transação calculando o SHA-256 completo dela
	hash := sha256.Sum256([]byte(dataToSign + sig))
	tx.ID = hex.EncodeToString(hash[:])

	return nil
}

// IsValid verifica se a assinatura da transação bate com a chave pública do remetente
func (tx *Transaction) IsValid() bool {
	// Se for uma emissão inicial do sistema (Bloco Gênese), podemos aceitar como válida
	if tx.Sender == "SISTEMA_CONSÓRCIO" {
		return true
	}

	dataToVerify := tx.GetSignableString()
	return VerifySignature(tx.Sender, dataToVerify, tx.Signature)
}
