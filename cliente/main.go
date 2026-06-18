package main

import (
	"fmt"
	"pbl/core"
)

func main() {
	// Exemplo de criação de uma transação
	transaction := core.Transaction{
		ID:        "tx123",
		Sender:    "sender_public_key",
		Receiver:  "receiver_public_key",
		Amount:    100.0,
		Payload:   "Laudo da missão",
		Signature: "signature",
	}

	fmt.Println("Transação criada:", transaction)
}
