# Form Builder

Aplicação para criar formulários, publicá-los como link e receber respostas.

Um usuário autenticado monta um formulário com múltiplos campos numa área
administrativa e o publica. A aplicação gera uma URL pública que qualquer
pessoa pode abrir e preencher **sem ter conta**. As respostas ficam
persistidas e são consultáveis por quem criou o formulário.

Seis tipos de campo: texto curto, texto longo, e-mail, número, seleção e caixa
de seleção — cada um com suas próprias regras de validação, aplicadas no
servidor.

---

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.24+, chi, cobra, viper |
| Banco | SQLite (`modernc.org/sqlite`, Go puro, sem cgo) |
| Migrations | goose, embutidas no binário |
| Contrato | OpenAPI 3 — gera o servidor Go e o client TypeScript |
| Frontend | React 19, TypeScript, Vite |
| Interface | MUI |
| Rotas | React Router |
| Estado remoto | TanStack Query |
| Autenticação | JWT em cookie HttpOnly, e-mail/senha ou Google |
| Tarefas | go-task |

---

## Pré-requisitos

- [Go 1.24+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org)
- [go-task](https://taskfile.dev): `go install github.com/go-task/task/v3/cmd/task@latest`

Nada além disso. Sem Docker, sem banco para instalar, sem compilador C. Os
geradores de código já vêm declarados no projeto.

> `go install` coloca o binário em `$(go env GOPATH)/bin` — confirme que esse
> diretório está no seu `PATH`.

---

## Rodando

```bash
task setup      # dependências + cria backend/config.yaml com um segredo próprio
task dev:api    # backend em http://localhost:8080
task dev:web    # frontend em http://localhost:5173   (outro terminal)
```

Abra <http://localhost:5173> e crie uma conta.

**Não há arquivo para editar antes de rodar.** O `task setup` gera o
`backend/config.yaml` e o banco é criado no primeiro boot, já com as migrations
aplicadas.

---

## Testando o fluxo

1. Crie uma conta em <http://localhost:5173>
2. **Criar formulário** → dê um título
3. **Adicionar campo** algumas vezes; troque os tipos e use ↑ ↓ para reordenar
4. **Salvar campos** → a pré-visualização aparece abaixo
5. **Publicar** → surge o link público
6. Abra o link **numa aba anônima** — é o que o visitante vê, sem login
7. Envie o formulário em branco: o backend recusa e aponta cada campo
8. Preencha e envie
9. Volte ao admin → **Respostas**

---

## Configuração

Nada precisa ser configurado para rodar — o `task setup` cuida disso. Para
mudar alguma coisa, edite `backend/config.yaml`. Toda chave também aceita
variável de ambiente com prefixo `FB_` e flag de linha de comando, nesta
precedência: **flag > env > arquivo > default**.

```bash
server run --address=localhost:9000
FB_DATABASE_PATH=./outro.db server run
```

O banco é um arquivo SQLite criado no primeiro boot, e as migrations rodam
sozinhas junto. `task migrate` existe para aplicá-las sem subir a API.

**Login com Google é opcional** — sem credenciais, a aplicação funciona inteira
com e-mail e senha. Para habilitar, crie um *ID do cliente OAuth* do tipo
Aplicativo da Web em <https://console.cloud.google.com/apis/credentials>, com
esta URI de redirecionamento:

```
http://localhost:8080/api/v1/auth/google/callback
```

e preencha `auth.google.client_id` e `auth.google.client_secret` em
`backend/config.yaml`, que está no `.gitignore`.

---

## Gerando a especificação OpenAPI e o client TypeScript

```bash
task generate   # gera os dois lados
task verify     # falha se o gerado estiver fora de sincronia com a spec
```

[`api/openapi.yaml`](api/openapi.yaml) é a fonte da verdade. Dois geradores
leem esse arquivo:

- **`oapi-codegen`** → interface e tipos Go em `backend/internal/httpapi/gen/`
- **`@hey-api/openapi-ts`** → tipos, client e hooks do TanStack Query em
  `frontend/src/api/gen/`

Ao mudar a API: edite `api/openapi.yaml`, rode `task generate`. O Go para de
compilar até o handler novo existir, e o TypeScript para de tipar se o front
usar o campo antigo.

O código gerado é versionado de propósito, para mudança de contrato aparecer no
diff. `task verify` regera tudo e falha se o resultado diferir do commitado.

O servidor também expõe a spec que o gerou, em <http://localhost:8080/openapi.json>.

---

## Comandos

```bash
task --list     # lista tudo
task setup      # dependências e configuração inicial
task dev:api    # backend
task dev:web    # frontend
task migrate    # migrations
task generate   # regera Go e TypeScript a partir da spec
task verify     # confere se o gerado está em sincronia
task test       # testes do backend
task build      # compila para ./bin/server
```

---

## Decisões de arquitetura

- **Spec-first.** O `openapi.yaml` é escrito à mão e gera os dois lados. Com o
  *strict server* do `oapi-codegen`, handler fora do contrato não compila.
- **SQLite via driver em Go puro**, para rodar igual em Windows, macOS e Linux
  sem Docker nem toolchain C. Os repositórios estão atrás de interfaces, então
  trocar de banco fica contido numa camada.
- **Sessão em JWT dentro de cookie `HttpOnly`**, inacessível ao JavaScript.
  Dispensa tabela de sessão; em troca, não tem revogação.
- **Autorização na assinatura do repositório.** Não existe `GetForm(id)`, só
  `GetOwnedBy(id, ownerID)` — não dá para esquecer a checagem de dono porque
  não dá para chamar o método sem ele.
- **`PUT` na lista inteira de campos**, em vez de CRUD por campo: o builder é
  uma unidade de edição, e o array recebido é a verdade.
- **Estrutura travada após publicar**, para não invalidar respostas já
  recebidas.
- **O rótulo do campo é copiado para a resposta** no momento do envio, então
  renomear uma pergunta depois não reescreve o passado.
- **Validação de submissão isolada**, sem dependência de HTTP ou banco. É o
  único pacote com suíte de testes.
- **Proxy do Vite em desenvolvimento**, então navegador e API são a mesma
  origem e não existe CORS.

---

## Limitações conhecidas

Tudo que o desafio pede está implementado. O que segue são limites de escopo,
assumidos conscientemente:

- **Sessão sem revogação.** O logout apaga o cookie do navegador, mas o token
  continua válido até expirar. Sessão em banco seria o próximo passo.
- **Editar um formulário publicado exige despublicá-lo**, e o link fica fora do
  ar nesse intervalo. A solução completa é versionar a definição, com cada
  resposta apontando para a versão que o visitante viu.
- **Sem rate limiting** nos endpoints de autenticação nem no endpoint público.
  Um limiter em memória sumiria no restart e não valeria com múltiplas
  instâncias; o passo real é Redis.
- **SQLite aceita um escritor por vez.** Suficiente aqui, mas escala apenas
  verticalmente.
- **Respostas armazenadas como texto.** Agregação numérica exigiria CAST.
- **Sem exportação CSV** das respostas.
- **Sem paginação na lista de formulários** (as respostas paginam).
- **Testes concentrados no validador.** Sem testes de integração HTTP e sem
  testes de frontend.
- **Sem verificação de e-mail no cadastro e sem recuperação de senha.**
- **Faltam tipos de campo**: upload de arquivo, múltipla escolha e lógica
  condicional entre campos.
- **Sem CAPTCHA** no envio público.
- **Bundle do frontend sem code splitting**, e sem Content-Security-Policy
  completa.
