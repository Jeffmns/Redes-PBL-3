package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// GenerateKeyPair gera um par de chaves (pública e privada) usando Ed25519.
// Retorna as duas chaves formatadas como strings hexadecimais limpas.
func GenerateKeyPair() (string, string, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Convertemos para formato Hexadecimal para facilitar o envio via JSON/Rede
	pubHex := hex.EncodeToString(pubKey)
	privHex := hex.EncodeToString(privKey)

	return pubHex, privHex, nil
}

// SignData assina uma string de dados textuais usando a chave privada hexadecimal.
// Retorna a assinatura resultante formatada como string hexadecimal.
func SignData(data string, privateKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", err
	}

	if len(privKeyBytes) != ed25519.PrivateKeySize {
		return "", errors.New("tamanho de chave privada Ed25519 inválido")
	}

	privKey := ed25519.PrivateKey(privKeyBytes)

	// Gera a assinatura digital baseada nos bytes da string de dados
	signature := ed25519.Sign(privKey, []byte(data))

	return hex.EncodeToString(signature), nil
}

// VerifySignature valida se uma assinatura digital é legítima para o par dado de dados + chave pública.
func VerifySignature(publicKeyHex string, data string, signatureHex string) bool {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	// Validação básica de segurança de tamanhos antes de processar
	if len(pubKeyBytes) != ed25519.PublicKeySize || len(sigBytes) != ed25519.SignatureSize {
		return false
	}

	pubKey := ed25519.PublicKey(pubKeyBytes)

	// Executa a verificação criptográfica pura
	return ed25519.Verify(pubKey, []byte(data), sigBytes)
}
