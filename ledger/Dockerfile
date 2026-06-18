# Usamos a imagem oficial do Go
FROM golang:1.25-alpine

# Definimos o diretório de trabalho principal
WORKDIR /workspace

# Copiamos TODO o repositório para dentro do contêiner,
COPY . .

# Entramos na pasta do ledger para compilar
WORKDIR /workspace/ledger

# Baixamos as dependências (CometBFT, etc) e compilamos o executável
RUN go mod download
RUN go build -o /usr/local/bin/ledger-app main.go

# Expomos a porta TCP que o CometBFT usará para se conectar via ABCI
EXPOSE 26658

# Comando para iniciar a nossa aplicação
CMD ["ledger-app"]