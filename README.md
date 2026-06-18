# Redes-PBL-3 [wip]

## Comandos:

Gerar carteira no cliente: docker exec redes-pbl-3-cliente-cli-1 go run cliente/main.go gerar-carteira
Realizar transação: docker exec redes-pbl-3-cliente-cli-1 go run cliente/main.go pagar-escolta -pub="" -priv="" -valor=50 -laudo="Missão inicial do Consórcio"

Substituindo pub e priv pelas chaves geradas na etapa de geração da carteira