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
task setup      # dependências + cria backend/config.yaml com um segredo próprio
task dev:api    # backend em localhost:8080
task dev:web    # frontend em localhost:5173  (outro terminal)
```

**Não há arquivo para editar antes de rodar.** O `task setup` gera o
`backend/config.yaml` a partir do template embutido, já com um segredo de
sessão criado por `crypto/rand` — sem depender de `openssl`, que não vem no
Windows. Rodar `task setup` de novo **não** troca o segredo (isso invalidaria
as sessões ativas); para forçar, `server init-config --force`.

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

O `task setup` já cria o `backend/config.yaml` para você. Para criar um em
outro lugar, ou recriar:

```bash
cd backend
go run ./cmd/server init-config                 # cria ./config.yaml
go run ./cmd/server init-config --out=/tmp/x.yaml
go run ./cmd/server init-config --force         # sobrescreve, invalidando sessões
```

`config.yaml` está no `.gitignore` — **nenhum segredo é versionado**. O
template fica em `backend/internal/config/config.example.yaml`, embutido no
binário, e traz um placeholder que o servidor explicitamente recusa.

Sem `--config`, o servidor procura um `config.yaml` no diretório atual e segue
normalmente se não encontrar. **Com** `--config` apontando para um arquivo
inexistente, ele falha no boot: se você pediu um arquivo específico, subir em
silêncio com os valores padrão esconderia o problema.

### Chaves disponíveis

| Chave (YAML) | Variável de ambiente | Flag | Default |
|---|---|---|---|
| `address` | `FB_ADDRESS` | `--address` | `localhost:8080` |
| `public_base_url` | `FB_PUBLIC_BASE_URL` | `--public-base-url` | `http://localhost:5173` |
| `database.path` | `FB_DATABASE_PATH` | `--db-path` | `./formbuilder.db` |
| `auth.jwt_secret` | `FB_AUTH_JWT_SECRET` | `--jwt-secret` | — (obrigatório) |
| `auth.session_ttl` | `FB_AUTH_SESSION_TTL` | `--session-ttl` | `24h` |

**`auth.jwt_secret` é obrigatório e o servidor não sobe sem ele.** Recusa
também um segredo com menos de 32 bytes e o valor de exemplo do template.
Falhar no boot é deliberado: a alternativa é o servidor rodar meses assinando
sessões com um segredo publicado no repositório, e ninguém perceber.

Isso não custa nada a quem clona, porque o `task setup` já gera um segredo
próprio — a exigência protege sem criar passo manual.

O prefixo é sempre `FB_`, e chaves aninhadas trocam `.` por `_`
(`database.path` → `FB_DATABASE_PATH`).

Configuração inválida derruba o boot com mensagem explícita, em vez de virar
comportamento estranho na hora da requisição.

---

## Configurando o banco e executando migrations

**Não há banco para instalar nem container para subir.** A persistência é
SQLite através de [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite),
um driver escrito em Go puro: sem cgo, sem compilador C, sem Docker. É o que
faz o projeto rodar igual em Windows, macOS e Linux, que é o que o desafio pede.

O banco é um arquivo, criado no primeiro uso:

```bash
task migrate                                  # aplica as migrations
server migrate --db-path=./outro.db           # ou apontando para outro arquivo
```

O comando é idempotente — rodar duas vezes seguidas não faz nada na segunda.
**O servidor também aplica as migrations no boot**, então um clone limpo sobe
com `task dev:api` sem passo manual. `server migrate` existe para preparar o
banco sem subir a API.

As migrations vivem em `backend/internal/storage/sqlite/migrations/` e são
embutidas no binário com `embed.FS`: o executável não depende de encontrar os
arquivos `.sql` no disco. O versionamento é do
[goose](https://github.com/pressly/goose), que registra o estado na tabela
`goose_db_version`.

### Os pragmas, e por que estão no DSN

```
file:formbuilder.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
```

`foreign_keys` é **desligado por padrão** no SQLite, por compatibilidade
retroativa — sem ele, todos os `ON DELETE CASCADE` do schema seriam
decorativos. `journal_mode=WAL` permite leitores concorrentes com um escritor.
`busy_timeout` faz o escritor esperar em vez de estourar `database is locked`
na hora.

Eles estão no **DSN** e não num `PRAGMA` executado depois porque valem **por
conexão**, e o `database/sql` mantém um pool: um comando avulso configuraria
apenas a primeira conexão, e o bug apareceria de forma intermitente. No DSN, o
driver aplica em toda conexão nova.

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

## Autenticação

Sessão em **JWT HS256 dentro de um cookie `HttpOnly`**.

```bash
curl -c j.txt -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana","email":"ana@exemplo.com","password":"uma-senha-longa"}'

curl -b j.txt localhost:8080/api/v1/auth/me
```

### As decisões, e o porquê de cada uma

**Cookie `HttpOnly` em vez de token no `localStorage`.** O `localStorage` é
legível por qualquer JavaScript da página — uma XSS, minha ou de uma
dependência, entrega o token. Com `HttpOnly`, o JavaScript não alcança o
cookie. O custo é CSRF, coberto por `SameSite=Lax`. O trade-off honesto: o JWT
me poupa a tabela de sessão mas **não tem revogação** — o logout apaga o cookie
do navegador, e um token vazado vale até expirar. Por isso o TTL é curto.
Sessão opaca em banco é o passo seguinte.

**`Secure` no cookie é condicional.** Ligado em `http` de desenvolvimento, o
navegador descartaria o cookie **em silêncio** e o login pareceria quebrado sem
nenhum erro visível. Ele acompanha o esquema do `public_base_url`.

**`jwt.WithValidMethods` no parse.** A biblioteca entrega o token já parseado e
deixa você devolver a chave. Devolvendo o segredo sem olhar o header `alg`,
quem valida o algoritmo é o atacante: ele manda `alg: none` ou troca HMAC por
RSA. São duas linhas para fechar o abuso de JWT mais explorado que existe.

**O payload do JWT tem só o ID do usuário.** JWT é **assinado, não
criptografado**: qualquer um faz base64 do payload e lê. A assinatura garante
que ninguém alterou, não que ninguém leu.

**bcrypt com a senha limitada a 72 bytes.** O bcrypt ignora tudo depois do byte
72 **sem avisar**. Sem validar isso, uma senha de 100 caracteres é secretamente
uma de 72, e duas senhas diferentes com o mesmo prefixo produzem o mesmo hash.
Recuso na entrada em vez de truncar em silêncio.

**Mínimo de 8 caracteres, sem exigir símbolo ou maiúscula.** Regras de
complexidade empurram o usuário para `Senha@123` — previsível e curta.
Comprimento é o que importa (NIST 800-63B).

**E-mail inexistente e senha errada devolvem a mesma resposta — e gastam o
mesmo tempo.** A mensagem única impede que o login vire uma API de consulta de
contas. Só que sem usuário eu retornaria em microssegundos, contra ~60ms
comparando um hash real: a diferença é mensurável pela rede, o mesmo vazamento
por cronômetro. Então o caminho "usuário não encontrado" roda um bcrypt
descartável contra um hash fixo, só para pagar o mesmo custo.

**A proteção das rotas é derivada da spec.** O `openapi.yaml` declara
`cookieAuth` globalmente; uma rota pública precisa dizer `security: []`. O
middleware **lê essa declaração da spec embutida no boot** e monta a lista de
operações abertas. Nenhuma lista de rotas protegidas é mantida à mão, então
rota nova nasce protegida e abrir uma é uma linha visível no diff da spec.

**`Cache-Control: no-store` nas respostas autenticadas**, para o botão voltar
não recuperar dado privado depois do logout.

### O que não tem, conscientemente

- Rate limiting nos endpoints de autenticação (previsto, ainda não implementado)
- Bloqueio de conta após N tentativas — mal feito, vira DoS de conta
- 2FA, verificação de e-mail no cadastro e recuperação de senha
- `/auth/register` revela que um e-mail já existe (409). Trade-off aceito por
  UX numa área administrativa fechada

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
