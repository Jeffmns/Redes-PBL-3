# Usamos a imagem oficial do Go
FROM golang:alpine

# Definimos o diretório de trabalho principal
WORKDIR /workspace

# Copiamos TODO o repositório para o contêiner
COPY . .

# Entramos na pasta do cliente para compilar
WORKDIR /workspace/cliente

# Sincronizamos as dependências (e a biblioteca core local)
RUN go mod tidy

# Compilamos o binário do cliente
RUN go build -o /usr/local/bin/cliente-app main.go

# Como este contêiner serve apenas para enviarmos comandos manualmente (CLI),
# mantemo-lo vivo no background indefinidamente.
CMD ["tail", "-f", "/dev/null"]