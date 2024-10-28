#!/bin/bash

# Nome do serviço Docker Compose
SERVICE_NAME="othello"

# Verificar se já existe um processo do Go em execução
if pgrep -x "go" > /dev/null; then
  echo "Já existe um processo do Go em execução. Por favor, finalize-o antes de continuar."
  exit 1
fi

# Parar e remover contêiner existente com o mesmo nome (se houver)
if [ "$(docker compose ps -q $SERVICE_NAME)" ]; then
  echo "Parando e removendo contêiner existente..."
  docker compose down
fi

# Construir e rodar o contêiner Docker usando Docker Compose
echo "Construindo e rodando o contêiner Docker..."
docker compose up --build -d

# Verificar se o contêiner foi iniciado corretamente
if [ $? -ne 0 ]; then
  echo "Erro ao iniciar o contêiner Docker"
  exit 1
fi

# Esperar alguns segundos para garantir que o servidor esteja pronto
sleep 3

# Abrir duas abas do navegador apontando para o servidor
echo "Abrindo duas abas do navegador..."
if which xdg-open > /dev/null; then
  xdg-open http://localhost:8080/
  xdg-open http://localhost:8080/
elif which open > /dev/null; then
  open http://localhost:8080/
  open http://localhost:8080/
elif which start > /dev/null; then
  start http://localhost:8080/
  start http://localhost:8080/
else
  echo "Não foi possível detectar o comando para abrir o navegador."
fi

# Exibir os logs do contêiner
echo "Exibindo logs do servidor..."
docker compose logs -f $SERVICE_NAME