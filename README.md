# Form Builder

Aplicação de criação e publicação de formulários. Um usuário autenticado cria
formulários numa área administrativa, publica, e a aplicação gera um link
público que qualquer pessoa pode preencher sem autenticação. As respostas ficam
persistidas e são consultáveis pelo administrador.

Backend em Go, frontend em React + TypeScript, contrato em OpenAPI 3 que gera os
dois lados.

> **Status:** em construção. Este README cresce a cada fase entregue; hoje ele
> cobre até o esqueleto do backend.

---

## Pré-requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.22+ | backend |
| [go-task](https://taskfile.dev) | 3.x | entrypoint de todos os comandos |

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

`go install` coloca o binário em `$(go env GOPATH)/bin` — garanta que esse
diretório está no `PATH`.

Não é necessário Docker, compilador C nem nenhum serviço externo: o banco será
SQLite através de um driver escrito em Go puro.

---

## Executar o backend

```bash
task setup      # baixa as dependências
task dev:api    # sobe o servidor em localhost:8080
```

Verificando:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Outros comandos:

```bash
task --list     # lista tudo
task build      # compila para ./bin/server
task test       # go test ./...
task vet        # análise estática
```

---

## Configurar a aplicação

A configuração vem de quatro fontes, nesta ordem de precedência:

**flag > variável de ambiente > arquivo > default**

Ou seja, as três formas que o desafio pede funcionam e podem ser combinadas:

```bash
# arquivo
server run --config=./config.yaml

# flag
server run --address=localhost:9000

# variável de ambiente
FB_ADDRESS=localhost:9000 server run
```

Para criar seu arquivo de configuração:

```bash
cd backend
cp config.example.yaml config.yaml
```

`config.yaml` está no `.gitignore` — só o `.example` é versionado, e ele nunca
contém segredo real.

Sem `--config`, o servidor procura um `config.yaml` no diretório atual e segue
normalmente se não encontrar. **Com** `--config` apontando para um arquivo que
não existe, ele falha na hora: se você pediu um arquivo específico, subir em
silêncio com os valores padrão esconderia o problema.

### Chaves disponíveis

| Chave (YAML) | Variável de ambiente | Flag | Default |
|---|---|---|---|
| `address` | `FB_ADDRESS` | `--address` | `localhost:8080` |
| `public_base_url` | `FB_PUBLIC_BASE_URL` | `--public-base-url` | `http://localhost:5173` |

O prefixo é sempre `FB_`, e chaves aninhadas trocam `.` por `_`
(`auth.jwt_secret` → `FB_AUTH_JWT_SECRET`).

`public_base_url` é a origem do frontend: é a partir dela que o link público de
um formulário publicado é montado, então o link sai correto em qualquer
ambiente sem recompilar nada.

Configuração inválida derruba o boot com mensagem explícita, em vez de virar
comportamento estranho na hora da requisição.

---

## Decisões de arquitetura

Registradas aqui conforme são tomadas, com o porquê e a alternativa descartada.

**`main` só resolve sinais do sistema operacional.** O pacote `main` não pode
ser importado por ninguém, então lógica que cai lá é lógica que não tem como
ser testada. `cmd/server/main.go` traduz SIGINT/SIGTERM em cancelamento de
contexto e delega para `internal/cli`.

**`internal/` é intencional.** Nada fora do módulo consegue importar as camadas
internas — é o compilador garantindo a fronteira, não convenção de nome de
pasta.

**Timeouts explícitos no `http.Server`.** O zero value de `http.Server` em Go
não tem timeout nenhum: uma conexão que envia o cabeçalho byte a byte segura
uma goroutine para sempre (Slowloris). São quatro linhas que a maioria dos
projetos Go esquece, e por isso entraram no esqueleto e não "depois".

**Desligamento gracioso.** Ctrl+C termina as requisições em andamento antes de
fechar, para não cortar uma submissão no meio da transação.

**Cabeçalhos de segurança em toda resposta:** `X-Content-Type-Options: nosniff`,
`X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`.

**`/health` fora de `/api/v1`.** É endpoint de operação, não parte do contrato
do produto.

---

## Estrutura

```
.
├── Taskfile.yml              # entrypoint único de DX
├── api/                      # openapi.yaml — fonte da verdade (em breve)
└── backend/
    ├── config.example.yaml
    ├── cmd/server/main.go    # só sinais; delega para internal/cli
    └── internal/
        ├── cli/              # cobra: root, run
        ├── config/           # viper: default → arquivo → env → flag
        └── httpapi/          # chi, middlewares, servidor
```
