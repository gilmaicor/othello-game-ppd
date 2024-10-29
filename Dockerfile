# Etapa 1: Build da aplicação Go
FROM golang:1.21-alpine AS builder

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos go.mod e go.sum e baixar dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante dos arquivos do projeto
COPY . .

# Compilar a aplicação
RUN go build -o othello main.go

# Etapa 2: Configuração do ambiente de execução
FROM alpine:latest

# Instalar dependências necessárias
RUN apk --no-cache add ca-certificates

# Definir diretório de trabalho
WORKDIR /root/

# Copiar o binário compilado da etapa anterior
COPY --from=builder /app .

# Copiar arquivos estáticos
COPY --from=builder /app/static ./static

# Expor a porta que a aplicação irá rodar
EXPOSE 8080

# Definir comando de inicialização
CMD ["./othello"]