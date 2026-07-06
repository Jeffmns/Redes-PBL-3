# Usamos a imagem oficial do Go
FROM golang:alpine

# Definimos o diretório de trabalho principal
WORKDIR /workspace

# Copiamos TODO o repositório para dentro do contêiner
COPY . .

# Entramos na pasta do ledger para compilar
WORKDIR /workspace/ledger

# Usamos tidy em vez de download para sincronizar dependências automaticamente
RUN go mod tidy

# Compilamos o executável
RUN go build -o /usr/local/bin/ledger-app main.go

# Expomos as portas usadas pelo ABCI e pela API REST do ledger
EXPOSE 26658 8080

# Comando para iniciar a nossa aplicação
CMD ["ledger-app"]