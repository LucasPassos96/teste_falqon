# Form Builder

Aplicação de criação e publicação de formulários. Um usuário autenticado cria
formulários numa área administrativa, publica, e a aplicação gera um link
público que qualquer pessoa pode preencher sem autenticação. As respostas ficam
persistidas e são consultáveis pelo administrador.

Backend em Go, frontend em React + TypeScript, e um contrato OpenAPI 3 que gera
os dois lados.

> **Status:** em construção. Este README cresce a cada fase entregue; hoje ele
> cobre o esqueleto do backend e o pipeline de geração de código.

---

## Pré-requisitos

| Ferramenta | Versão | Para quê |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.24+ | backend |
| [Node.js](https://nodejs.org) | 20+ | frontend |
| [go-task](https://taskfile.dev) | 3.x | entrypoint de todos os comandos |

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

`go install` coloca o binário em `$(go env GOPATH)/bin` — garanta que esse
diretório está no `PATH`.

Não é necessário Docker nem compilador C. O gerador de código Go está fixado
como `tool` no `go.mod`, então **não há ferramenta para instalar à parte**.

---

## Executando

```bash
task setup      # dependências do backend e do frontend
task dev:api    # backend em localhost:8080
task dev:web    # frontend em localhost:5173  (outro terminal)
```

Abra <http://localhost:5173>. A página consulta a API e mostra a resposta.

Verificando só o backend:

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}

curl http://localhost:8080/openapi.json
# a especificação servida pelo próprio binário
```

Outros comandos:

```bash
task --list     # lista tudo
task generate   # regera o servidor Go e o client TypeScript
task verify     # falha se o gerado estiver fora de sincronia com a spec
task build      # compila para ./bin/server
task vet        # análise estática
```

**Sem CORS em desenvolvimento:** o Vite faz proxy de `/api` para
`localhost:8080`, então navegador e API são a mesma origem. Isso elimina toda
uma classe de problema de preflight e de cookie que apareceria mais adiante,
na autenticação.

---

## Configurando a aplicação

A configuração vem de quatro fontes, nesta ordem de precedência:

**flag > variável de ambiente > arquivo > default**

As três formas que o desafio cita funcionam e podem ser combinadas:

```bash
server run --config=./config.yaml        # arquivo
server run --address=localhost:9000      # flag
FB_ADDRESS=localhost:9000 server run     # variável de ambiente
```

Para criar seu arquivo de configuração:

```bash
cd backend
cp config.example.yaml config.yaml
```

`config.yaml` está no `.gitignore` — só o `.example` é versionado, e ele nunca
contém segredo real.

Sem `--config`, o servidor procura um `config.yaml` no diretório atual e segue
normalmente se não encontrar. **Com** `--config` apontando para um arquivo
inexistente, ele falha no boot: se você pediu um arquivo específico, subir em
silêncio com os valores padrão esconderia o problema.

### Chaves disponíveis

| Chave (YAML) | Variável de ambiente | Flag | Default |
|---|---|---|---|
| `address` | `FB_ADDRESS` | `--address` | `localhost:8080` |
| `public_base_url` | `FB_PUBLIC_BASE_URL` | `--public-base-url` | `http://localhost:5173` |

O prefixo é sempre `FB_`, e chaves aninhadas trocam `.` por `_`
(`auth.jwt_secret` → `FB_AUTH_JWT_SECRET`).

Configuração inválida derruba o boot com mensagem explícita, em vez de virar
comportamento estranho na hora da requisição.

---

## Gerando a especificação OpenAPI e o client TypeScript

```bash
task generate
```

Um comando gera os dois lados.

### Como funciona

[`api/openapi.yaml`](api/openapi.yaml) é a **fonte da verdade**. Dois geradores
leem esse arquivo:

```
                      api/openapi.yaml
                     /                \
          oapi-codegen                @hey-api/openapi-ts
               ↓                              ↓
   backend/internal/httpapi/gen        frontend/src/api/gen
   (interface + tipos Go)              (tipos + client + hooks Query)
```

O fluxo de trabalho ao mudar a API:

1. Edita `api/openapi.yaml`
2. `task generate`
3. O Go **para de compilar** até o handler novo existir
4. O TypeScript **para de tipar** se o front usar o campo antigo

Isso não é teoria — dá para reproduzir. Adicione um endpoint na spec, rode
`task generate` e compile:

```
*API does not implement gen.StrictServerInterface (missing method GetPing)
```

Renomeie um campo de um schema e rode `tsc`:

```
error TS2339: Property 'status' does not exist on type 'Health'.
```

Divergir da spec é erro de compilação, não bug em produção.

### Por que a spec é escrita à mão e não extraída do código

O desafio aceita as duas formas: especificação *"gerada automaticamente a
partir do backend **ou** parte de um processo de geração integrado ao código"*.
A segunda foi escolhida de propósito.

Com anotações no código, a spec **descreve** o que o backend faz: se o handler
está errado, a spec documenta o erro fielmente. Com spec-first e o
`strict-server` do `oapi-codegen`, o Go implementa uma interface gerada — um
handler fora do contrato não compila. E o mesmo arquivo alimenta o client
TypeScript, então os dois lados derivam da mesma fonte e não têm como divergir.

O custo é escrever YAML à mão, que é verboso. Num time com API mudando rápido,
code-first reduz atrito, e aí a compensação seria teste de contrato.

### O código gerado é versionado

`backend/internal/httpapi/gen/` e `frontend/src/api/gen/` estão no repositório
de propósito: mudança de contrato aparece no diff do PR, revisável, em vez de
acontecer invisível no build. `task verify` regera tudo e falha se o resultado
diferir do que está commitado — pega quem editou código gerado à mão ou
esqueceu de rodar `task generate`. É o que roda em CI.

### As ferramentas não precisam ser instaladas

O `oapi-codegen` está fixado no `go.mod` com a diretiva `tool`, então
`go tool oapi-codegen` usa a versão exata que o projeto declara. O
`@hey-api/openapi-ts` vem do `package.json`. Ninguém precisa acertar a versão
de nada na mão, e não existe "na minha máquina funciona" por diferença de
ferramenta.

---

## Decisões de arquitetura

Registradas aqui conforme são tomadas, com o porquê e a alternativa descartada.

**Spec-first, com o contrato virando tipo nos dois lados.** Detalhado acima. É
a decisão central do projeto.

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

**A spec é servida pelo próprio binário** em `/openapi.json`, embutida em tempo
de geração. Garante que a especificação publicada é exatamente a que gerou
aquele executável.

---

## Estrutura

```
.
├── Taskfile.yml                  # entrypoint único de DX
├── api/
│   └── openapi.yaml              # ← FONTE DA VERDADE
├── backend/
│   ├── config.example.yaml
│   ├── oapi-codegen.yaml
│   ├── cmd/server/main.go        # só sinais; delega para internal/cli
│   └── internal/
│       ├── cli/                  # cobra: root, run
│       ├── config/               # viper: default → arquivo → env → flag
│       └── httpapi/
│           ├── gen/              # ← GERADO, versionado
│           ├── api.go            # implementa a interface gerada
│           ├── router.go
│           ├── middleware.go
│           └── server.go
└── frontend/
    ├── openapi-ts.config.ts
    ├── vite.config.ts            # proxy /api → :8080
    └── src/
        ├── api/gen/              # ← GERADO, versionado
        ├── main.tsx
        └── App.tsx
```
