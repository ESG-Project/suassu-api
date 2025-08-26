# Configuração da API

Este pacote gerencia toda a configuração da aplicação, incluindo variáveis de ambiente, logger e conexão com banco de dados.

## Estrutura

- **`config.go`**: Carregamento de variáveis de ambiente e validações
- **`log.go`**: Configuração do logger Zap
- **`db.go`**: Configuração da conexão PostgreSQL
- **`example.env`**: Exemplo de variáveis de ambiente

## Uso

### 1. Carregar Configuração

```go
import "github.com/ESG-Project/suassu-api/internal/config"

cfg, err := config.Load()
if err != nil {
    log.Fatal("failed to load config:", err)
}
```

### 2. Configurar Logger

```go
logger, err := config.BuildLogger(cfg)
if err != nil {
    log.Fatal("failed to build logger:", err)
}
defer logger.Sync()
```

### 3. Conectar ao Banco

```go
ctx := context.Background()
db, err := config.OpenPostgres(ctx, cfg)
if err != nil {
    logger.Fatal("failed to connect to database:", err)
}
defer db.Close()
```

## Variáveis de Ambiente

| Variável              | Padrão       | Descrição                                       |
| --------------------- | ------------ | ----------------------------------------------- |
| `APP_NAME`            | `suassu-api` | Nome da aplicação                               |
| `APP_ENV`             | `dev`        | Ambiente (`dev`, `staging`, `prod`)             |
| `HTTP_PORT`           | `8080`       | Porta do servidor HTTP                          |
| `DB_DSN`              | -            | String de conexão do banco (tem precedência)    |
| `DATABASE_URL`        | -            | String de conexão alternativa                   |
| `DB_MAX_OPEN_CONNS`   | `20`         | Máximo de conexões abertas                      |
| `DB_MAX_IDLE_CONNS`   | `10`         | Máximo de conexões ociosas                      |
| `DB_CONN_MAX_IDLE_MS` | `60000`      | Tempo máximo ocioso (ms)                        |
| `DB_CONN_MAX_LIFE_MS` | `300000`     | Tempo máximo de vida (ms)                       |
| `LOG_LEVEL`           | `info`       | Nível de log (`debug`, `info`, `warn`, `error`) |

## Validações

- **Produção**: `DB_DSN` ou `DATABASE_URL` é obrigatório
- **Porta HTTP**: Deve ser especificada
- **Valores inteiros**: Invalidos fazem fallback para padrões

## Exemplo de Configuração

Copie `example.env` para `.env` na **raiz do projeto** e ajuste os valores:

```bash
# Na raiz do projeto (suassu-api/)
cp internal/config/example.env .env

# Editar o .env com suas configurações
nano .env

# Carregar as variáveis no terminal
source .env
```

## Testes

Execute os testes com:

```bash
go test ./internal/config/... -v
```
