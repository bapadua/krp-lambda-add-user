# Kids Rewards Platform - Add User Lambda

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go)
![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-FF9900?logo=amazon-aws)
![Deploy](https://github.com/YOUR_USERNAME/krp-lambda-add-user/actions/workflows/deploy.yml/badge.svg)

Lambda Function para adicionar usuários na plataforma Kids Rewards.

## 📋 Funcionalidades

- ✅ Criar novos usuários no DynamoDB
- ✅ Validação de dados de entrada
- ✅ Geração automática de ID (UUID)
- ✅ Timestamp de criação
- ✅ Status inicial "active"

## 🏗️ Arquitetura

```
API Gateway / EventBridge
         ↓
   Lambda: add-user
         ↓
   DynamoDB: Users Table
```

## 📊 Schema do Usuário

### Request (Input)

```json
{
  "name": "João Silva",
  "phone_number": "+5511999999999",
  "email": "joao@example.com",
  "age": 10
}
```

**Campos obrigatórios:**
- `name` (string): Nome do usuário
- `phone_number` (string): Número de telefone

**Campos opcionais:**
- `email` (string): Email do usuário
- `age` (number): Idade do usuário

### Response (Output)

#### Sucesso (201)

```json
{
  "message": "User created successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "João Silva",
    "phone_number": "+5511999999999",
    "email": "joao@example.com",
    "age": 10,
    "status": "active",
    "created_at": "2025-11-17T12:00:00Z"
  }
}
```

#### Erro (400)

```json
{
  "error": "name is required"
}
```

#### Erro (500)

```json
{
  "error": "failed to save user: <error details>"
}
```

## 🚀 Estrutura do Projeto

```
krp-lambda-add-user/
├── main.go                 # Código principal da Lambda
├── go.mod                  # Dependências Go
├── go.sum                  # Checksum das dependências
├── .github/
│   └── workflows/
│       └── deploy.yml      # CI/CD Pipeline
├── .gitignore              # Arquivos ignorados
└── README.md               # Este arquivo
```

## 🛠️ Tecnologias

- **Runtime**: Go 1.21
- **AWS Services**:
  - Lambda (provided.al2023)
  - DynamoDB
  - S3 (artifacts)
  - CloudWatch Logs
  - X-Ray (tracing)

## 📦 Dependências

```go
github.com/aws/aws-lambda-go       v1.47.0
github.com/aws/aws-sdk-go-v2       v1.30.3
github.com/aws/aws-sdk-go-v2/config v1.27.27
github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.4
github.com/google/uuid              v1.6.0
```

## 🔧 Desenvolvimento Local

### Pré-requisitos

- Go 1.21+
- AWS CLI configurado
- Credenciais AWS

### Instalar Dependências

```bash
go mod download
```

### Build

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o bootstrap main.go
```

### Executar Localmente

```bash
# Usando AWS SAM
sam local invoke -e event.json

# Ou usando Lambda Runtime Interface Emulator
docker run -p 9000:8080 -e USERS_TABLE_NAME=your-table \
  -v "$PWD":/var/task public.ecr.aws/lambda/provided:al2023 \
  ./bootstrap
```

### Testes

```bash
go test -v ./...
```

## 🚀 Deploy

O deploy é automático via GitHub Actions quando há push na branch `main`.

### Deploy Manual

```bash
# Build
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o bootstrap main.go

# Package
zip lambda.zip bootstrap

# Upload para S3
aws s3 cp lambda.zip \
  s3://kids-rewards-platform-lambda-artifacts-prod/add-user/add-user.zip

# Update Lambda
aws lambda update-function-code \
  --function-name kids-rewards-platform-add-user-prod \
  --s3-bucket kids-rewards-platform-lambda-artifacts-prod \
  --s3-key add-user/add-user.zip
```

## 🔐 Variáveis de Ambiente

| Nome | Descrição | Obrigatória |
|------|-----------|-------------|
| `USERS_TABLE_NAME` | Nome da tabela DynamoDB | ✅ |

## 📊 Monitoramento

### CloudWatch Logs

```bash
aws logs tail /aws/lambda/kids-rewards-platform-add-user-prod --follow
```

### Métricas

- Invocations
- Duration
- Errors
- Throttles

### X-Ray Tracing

O X-Ray está habilitado para rastreamento de execução e análise de performance.

## 🐛 Troubleshooting

### Lambda não atualiza o código

```bash
# Verificar se o arquivo foi enviado
aws s3 ls s3://kids-rewards-platform-lambda-artifacts-prod/add-user/

# Forçar atualização
aws lambda update-function-code \
  --function-name kids-rewards-platform-add-user-prod \
  --s3-bucket kids-rewards-platform-lambda-artifacts-prod \
  --s3-key add-user/add-user.zip \
  --publish
```

### Erro de permissão no DynamoDB

Verifique se a IAM Role da Lambda tem permissões para:
- `dynamodb:PutItem`
- `dynamodb:GetItem`

### Erro no build

Certifique-se de usar as flags corretas:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0
```

## 📝 Estrutura da Tabela DynamoDB

### Users Table

| Campo | Tipo | Descrição | Key |
|-------|------|-----------|-----|
| id | String | UUID único | HASH |
| phone_number | String | Telefone | RANGE |
| name | String | Nome completo | GSI |
| email | String | Email (opcional) | - |
| age | Number | Idade (opcional) | - |
| status | String | Status (active/inactive) | GSI |
| created_at | String | Timestamp ISO8601 | - |

### Índices (GSI)

1. **name-index**: `name` (HASH) + `created_at` (RANGE)
2. **status-index**: `status` (HASH) + `created_at` (RANGE)

## 🔄 CI/CD Pipeline

O pipeline do GitHub Actions executa:

1. **Test**: Testes e validação de código
2. **Build**: Compilação para Linux AMD64
3. **Package**: Criação do ZIP
4. **Deploy**: Upload para S3 e atualização da Lambda

## 📞 Contato

- **Project**: Kids Rewards Platform
- **Environment**: PROD
- **Region**: us-east-1

## 📝 License

Proprietary - All rights reserved

---

**Última atualização**: 2025-11-17

