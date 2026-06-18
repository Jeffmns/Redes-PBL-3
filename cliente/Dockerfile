# Usamos a imagem oficial do Go
FROM golang:1.25-alpine

# Definimos o diretório de trabalho principal
WORKDIR /workspace

# Copiamos todo o projeto para dentro do contêiner
COPY . .

# Baixamos as dependências do core (se houver)
RUN go mod download