# Guia: Como Criar uma Lambda

## 📋 Pré-requisitos

- Infraestrutura base já deployada (S3, DynamoDB)
- Secret `AWS_ARN_ROLE_SECRET` configurado no GitHub
- Trust Policy do IAM Role atualizada

---

## 🚀 Passo 1: Criar Infraestrutura da Lambda

### 1.1. Edite `environments/prod/lambdas.tf`

Descomente e customize o módulo da Lambda:

```hcl
module "process_task_lambda" {
  source = "../../modules/lambda"

  project_name  = var.project_name
  function_name = "process-task"  # Nome da sua Lambda
  environment   = var.environment
  owner         = var.owner

  # Configurações Golang
  runtime     = "provided.al2023"
  handler     = "bootstrap"
  timeout     = 30
  memory_size = 512

  lambda_artifacts_bucket = module.lambda_artifacts.bucket_name

  # Variáveis de ambiente
  environment_variables = {
    USERS_TABLE_NAME = module.dynamodb.users_table_name
    TASKS_TABLE_NAME = module.dynamodb.tasks_table_name
  }

  log_retention_days = 30
  enable_xray        = true

  # Permissões DynamoDB
  custom_policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:Query",
        "dynamodb:Scan"
      ]
      Resource = [
        module.dynamodb.users_table_arn,
        module.dynamodb.tasks_table_arn,
        "${module.dynamodb.users_table_arn}/index/*",
        "${module.dynamodb.tasks_table_arn}/index/*"
      ]
    }]
  })
}
```

### 1.2. Descomente os Outputs

Em `environments/prod/outputs.tf`, descomente:

```hcl
output "process_task_lambda_arn" {
  description = "ARN da Lambda process-task"
  value       = module.process_task_lambda.function_arn
}
```

### 1.3. Deploy

```bash
cd environments/prod
terraform init
terraform plan
terraform apply
```

**Anote os outputs:**
- Bucket S3: `kids-rewards-platform-lambda-artifacts-prod`
- Lambda Name: `kids-rewards-platform-process-task-prod`

---

## 📦 Passo 2: Criar Repositório da Lambda

### 2.1. Criar Repositório no GitHub

```bash
gh repo create lambda-process-task --private --clone
cd lambda-process-task
```

### 2.2. Estrutura de Arquivos

```
lambda-process-task/
├── main.go
├── go.mod
├── go.sum
├── .github/
│   └── workflows/
│       └── deploy.yml
├── .gitignore
└── README.md
```

### 2.3. Código da Lambda (`main.go`)

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Request struct {
	UserID string `json:"userId"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

func HandleRequest(ctx context.Context, req Request) (Response, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return errorResponse(500, err.Error())
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	
	// Exemplo: Get item do DynamoDB
	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: req.UserID},
		},
	})
	if err != nil {
		return errorResponse(500, err.Error())
	}

	if result.Item == nil {
		return errorResponse(404, "User not found")
	}

	return Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"message": "Success"}`,
	}, nil
}

func errorResponse(code int, msg string) (Response, error) {
	body, _ := json.Marshal(map[string]string{"error": msg})
	return Response{
		StatusCode: code,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
```

### 2.4. Dependencies (`go.mod`)

```bash
go mod init lambda-process-task
go get github.com/aws/aws-lambda-go/lambda
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
```

### 2.5. .gitignore

```
bootstrap
*.zip
.DS_Store
```

---

## 🤖 Passo 3: Configurar CI/CD

### 3.1. Criar Workflow (`.github/workflows/deploy.yml`)

```yaml
name: Deploy Lambda

on:
  push:
    branches: [ main ]
  workflow_dispatch:

env:
  AWS_REGION: us-east-1
  LAMBDA_NAME: kids-rewards-platform-process-task-prod
  S3_BUCKET: kids-rewards-platform-lambda-artifacts-prod
  S3_KEY: process-task/process-task.zip
  GO_VERSION: '1.21'

permissions:
  id-token: write
  contents: read

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download Dependencies
        run: go mod download

      - name: Run Tests
        run: go test -v ./...

  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    needs: test
    environment: PROD
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build
        run: |
          GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
            go build -ldflags="-s -w" -o bootstrap main.go

      - name: Package
        run: zip lambda.zip bootstrap

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ARN_ROLE_SECRET }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Upload to S3
        run: |
          aws s3 cp lambda.zip s3://${{ env.S3_BUCKET }}/${{ env.S3_KEY }}

      - name: Update Lambda
        run: |
          aws lambda update-function-code \
            --function-name ${{ env.LAMBDA_NAME }} \
            --s3-bucket ${{ env.S3_BUCKET }} \
            --s3-key ${{ env.S3_KEY }}

      - name: Wait for Update
        run: |
          aws lambda wait function-updated \
            --function-name ${{ env.LAMBDA_NAME }}

      - name: Publish Version
        run: |
          aws lambda publish-version \
            --function-name ${{ env.LAMBDA_NAME }}
```

---

## 🔐 Passo 4: Configurar Secrets

### 4.1. Criar Environment no GitHub

```
Repo → Settings → Environments → New environment
Nome: PROD
```

### 4.2. Adicionar Secret

```
Environment: PROD → Add secret
Name: AWS_ARN_ROLE_SECRET
Value: <ARN do IAM Role>
```

### 4.3. Atualizar Trust Policy do IAM Role

Adicione o repositório da Lambda na Trust Policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::517888224370:oidc-provider/token.actions.githubusercontent.com"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
      },
      "StringLike": {
        "token.actions.githubusercontent.com:sub": [
          "repo:bapadua/b2p-infra-recompensa:*",
          "repo:bapadua/lambda-process-task:*"
        ]
      }
    }
  }]
}
```

---

## 🚀 Passo 5: Deploy

```bash
git add .
git commit -m "feat: initial lambda implementation"
git push origin main
```

GitHub Actions irá:
1. ✅ Rodar testes
2. ✅ Build do binário Go
3. ✅ Upload para S3
4. ✅ Atualizar Lambda Function

---

## 🧪 Passo 6: Testar

### Via AWS CLI

```bash
aws lambda invoke \
  --function-name kids-rewards-platform-process-task-prod \
  --payload '{"userId": "123"}' \
  response.json

cat response.json
```

### Ver Logs

```bash
aws logs tail /aws/lambda/kids-rewards-platform-process-task-prod --follow
```

---

## ✅ Checklist

- [ ] Módulo Lambda criado no Terraform
- [ ] Terraform apply executado
- [ ] Repositório GitHub criado
- [ ] Código Go implementado
- [ ] Workflow CI/CD configurado
- [ ] Environment PROD criado
- [ ] Secret AWS_ARN_ROLE_SECRET adicionado
- [ ] Trust Policy atualizada
- [ ] Push para main realizado
- [ ] Deploy bem-sucedido
- [ ] Lambda testada

---

## 🐛 Troubleshooting

### "Could not assume role"
- Verifique Trust Policy do IAM Role
- Verifique se o repositório está listado na Trust Policy
- Verifique se o secret está correto

### "Access Denied" ao fazer upload S3
- Verifique se a policy de deploy está anexada ao IAM Role
- ARN: `terraform output -raw lambda_deploy_policy_arn`

### Lambda não atualiza
```bash
# Forçar atualização
aws lambda update-function-code \
  --function-name kids-rewards-platform-process-task-prod \
  --s3-bucket kids-rewards-platform-lambda-artifacts-prod \
  --s3-key process-task/process-task.zip \
  --publish
```

### Build falha
- Certifique-se: `GOOS=linux GOARCH=amd64`
- Nome do binário: `bootstrap` (obrigatório)
- Go modules: `go mod tidy`

---

## 📝 Notas

- **Runtime:** `provided.al2023` para Go
- **Handler:** `bootstrap` (nome fixo)
- **Build:** Sempre cross-compile para Linux
- **S3:** Código versionado automaticamente
- **Rollback:** Use versões anteriores do S3

