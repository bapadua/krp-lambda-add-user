# 🚀 Próximos Passos - Lambda Add User

## ✅ O que já foi feito

1. ✅ Repositório Git inicializado
2. ✅ Código da Lambda criado (`main.go`)
3. ✅ Dependências configuradas (`go.mod`)
4. ✅ Workflow CI/CD criado (`.github/workflows/deploy.yml`)
5. ✅ README.md completo
6. ✅ Lambda adicionada na infraestrutura Terraform
7. ✅ Commit inicial realizado

## 📋 Próximos Passos

### 1. Criar Repositório no GitHub

```bash
# Opção 1: Usando GitHub CLI
gh repo create krp-lambda-add-user --private --source=. --remote=origin --push

# Opção 2: Manual
# 1. Acesse: https://github.com/new
# 2. Nome: krp-lambda-add-user
# 3. Visibilidade: Private
# 4. NÃO inicialize com README (já temos)
# 5. Clique em "Create repository"
```

### 2. Adicionar Remote e Push

Se criou manualmente, execute:

```bash
cd /d/src/rewards/krp-lambda-add-user
git remote add origin git@github.com:SEU_USERNAME/krp-lambda-add-user.git
git branch -M main
git push -u origin main
```

### 3. Configurar GitHub Secrets

#### 3.1. Criar Environment PROD

1. Acesse: `https://github.com/SEU_USERNAME/krp-lambda-add-user/settings/environments`
2. Clique em: **New environment**
3. Nome: `PROD`
4. Clique em: **Configure environment**

#### 3.2. Adicionar Secret

1. Na página do environment PROD
2. Clique em: **Add secret**
3. Nome: `AWS_ARN_ROLE_SECRET`
4. Valor: `arn:aws:iam::517888224370:role/github-actions-terraform-role`
5. Clique em: **Add secret**

### 4. Atualizar Trust Policy do IAM Role

Você precisa adicionar o novo repositório na Trust Policy do IAM Role:

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
          "repo:SEU_USERNAME/krp-lambda-add-user:*"
        ]
      }
    }
  }]
}
```

**Comando AWS CLI:**

```bash
# Salve a policy acima em um arquivo trust-policy.json
aws iam update-assume-role-policy \
  --role-name github-actions-terraform-role \
  --policy-document file://trust-policy.json
```

### 5. Deploy da Infraestrutura Lambda (Terraform)

```bash
cd /d/src/rewards/infra-aws/environments/prod

# Inicializar (se necessário)
terraform init

# Ver mudanças
terraform plan -var-file="terraform.tfvars"

# Aplicar
terraform apply -var-file="terraform.tfvars"
```

**Isso criará:**
- ✅ Lambda Function: `kids-rewards-platform-add-user-prod`
- ✅ IAM Role com permissões DynamoDB
- ✅ CloudWatch Log Group
- ✅ X-Ray tracing habilitado

### 6. Deploy do Código da Lambda

Após o Terraform criar a infraestrutura:

#### Opção 1: Via GitHub Actions (Recomendado)

```bash
# Faça qualquer alteração e push
git add .
git commit -m "chore: trigger deploy"
git push origin main
```

O workflow executará automaticamente! 🎉

#### Opção 2: Deploy Manual

```bash
cd /d/src/rewards/krp-lambda-add-user

# Baixar dependências
go mod download

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

# Aguardar atualização
aws lambda wait function-updated \
  --function-name kids-rewards-platform-add-user-prod

# Publicar versão
aws lambda publish-version \
  --function-name kids-rewards-platform-add-user-prod
```

### 7. Testar a Lambda

#### Via AWS CLI

```bash
# Criar evento de teste
cat > event.json << 'EOF'
{
  "name": "João Silva",
  "phone_number": "+5511999999999",
  "email": "joao@example.com",
  "age": 10
}
EOF

# Invocar Lambda
aws lambda invoke \
  --function-name kids-rewards-platform-add-user-prod \
  --payload file://event.json \
  --cli-binary-format raw-in-base64-out \
  response.json

# Ver resposta
cat response.json | jq
```

#### Via Console AWS

1. Acesse: https://console.aws.amazon.com/lambda/
2. Selecione: `kids-rewards-platform-add-user-prod`
3. Aba: **Test**
4. Event JSON:
   ```json
   {
     "name": "João Silva",
     "phone_number": "+5511999999999",
     "email": "joao@example.com",
     "age": 10
   }
   ```
5. Clique: **Test**

### 8. Verificar no DynamoDB

```bash
# Listar tabelas
aws dynamodb list-tables

# Scan users table
aws dynamodb scan \
  --table-name kids-rewards-platform-users-prod

# Get item específico (substitua o ID)
aws dynamodb get-item \
  --table-name kids-rewards-platform-users-prod \
  --key '{"id": {"S": "550e8400-e29b-41d4-a716-446655440000"}, "phone_number": {"S": "+5511999999999"}}'
```

### 9. Monitoramento

#### CloudWatch Logs

```bash
# Ver logs em tempo real
aws logs tail /aws/lambda/kids-rewards-platform-add-user-prod --follow

# Últimos 10 minutos
aws logs tail /aws/lambda/kids-rewards-platform-add-user-prod --since 10m
```

#### Métricas

```bash
# Ver invocações nas últimas 24h
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Invocations \
  --dimensions Name=FunctionName,Value=kids-rewards-platform-add-user-prod \
  --start-time $(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 3600 \
  --statistics Sum
```

## 📊 Arquitetura Final

```
GitHub Actions (CI/CD)
    ↓
S3 Bucket (Artifacts)
    ↓
Lambda: add-user
    ↓
DynamoDB: Users Table
```

## 🔐 Permissões IAM

A Lambda tem permissões para:
- ✅ `dynamodb:PutItem` - Criar usuários
- ✅ `dynamodb:GetItem` - Ler usuários
- ✅ `logs:CreateLogGroup` - Criar log groups
- ✅ `logs:CreateLogStream` - Criar log streams
- ✅ `logs:PutLogEvents` - Escrever logs
- ✅ `xray:PutTraceSegments` - X-Ray tracing
- ✅ `xray:PutTelemetryRecords` - X-Ray telemetry

## 📝 Checklist Final

- [ ] Repositório GitHub criado
- [ ] Remote origin configurado
- [ ] Push realizado
- [ ] Environment PROD criado
- [ ] Secret AWS_ARN_ROLE_SECRET configurado
- [ ] Trust Policy atualizada
- [ ] Terraform apply executado
- [ ] Lambda code deployed
- [ ] Teste executado com sucesso
- [ ] Usuário visível no DynamoDB

## 🎉 Pronto!

Sua Lambda está pronta para uso! 🚀

## 📞 Troubleshooting

### Erro: "Error assuming role"

- Verifique se a Trust Policy foi atualizada com o novo repositório
- Verifique se o secret AWS_ARN_ROLE_SECRET está correto
- Verifique se o environment PROD existe

### Erro: "Lambda not found"

- Execute `terraform apply` primeiro para criar a infraestrutura
- Verifique o nome da Lambda nos outputs do Terraform

### Erro: "Access Denied" no DynamoDB

- Verifique as permissões IAM da Lambda
- Certifique-se que a tabela existe

### Workflow não executa

- Verifique se está na branch `main`
- Verifique se o workflow file está em `.github/workflows/deploy.yml`
- Verifique os logs em: Actions → Deploy Lambda

---

**Última atualização**: 2025-11-17

