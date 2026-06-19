# Estágio 1: Build
FROM golang:alpine AS builder


WORKDIR /workspace

# Copiamos TODO o repositório para dentro do contêiner (Isto inclui a pasta core e a despachante)
COPY . .

# Entramos especificamente na pasta do despachante para compilar
WORKDIR /workspace/despachante

# Sincroniza e baixa automaticamente as dependências que estão faltando (MQTT, Websockets, etc)
RUN go mod tidy

# Compila o binário de forma estática
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# ---
# Estágio 2: Final (Imagem de execução leve)
FROM alpine:latest

# Instala certificados CA (necessário para chamadas HTTPS externas ao CometBFT)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia apenas o binário compilado do estágio anterior
COPY --from=builder /app/main .


EXPOSE 8080

CMD ["./main"]