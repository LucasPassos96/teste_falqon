-- +goose Up

-- Datas são TEXT em RFC3339: o SQLite não tem tipo de data, e texto ISO-8601
-- ordena corretamente em comparação lexicográfica.
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,   -- normalizado em minúsculas na entrada
    name          TEXT NOT NULL,
    password_hash TEXT,                   -- NULL para usuário que só entra pelo Google
    google_id     TEXT UNIQUE,            -- NULL para usuário que só entra por senha
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

CREATE TABLE forms (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL CHECK (status IN ('draft', 'published')),
    -- Gerado na primeira publicação e preservado ao despublicar/republicar:
    -- regenerar mataria links já distribuídos.
    public_slug  TEXT UNIQUE,
    published_at TEXT,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);

CREATE INDEX idx_forms_user_id ON forms (user_id);

CREATE TABLE form_fields (
    id        TEXT PRIMARY KEY,
    form_id   TEXT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    type      TEXT NOT NULL CHECK (
                  type IN ('short_text', 'long_text', 'email', 'number', 'select', 'checkbox')
              ),
    label     TEXT NOT NULL,
    help_text TEXT NOT NULL DEFAULT '',
    required  INTEGER NOT NULL DEFAULT 0,
    position  INTEGER NOT NULL,
    -- Regras específicas do tipo (options de select, min/max) num JSON só,
    -- validado na entrada. A alternativa era uma coluna por regra, esparsa e
    -- cheia de NULL.
    config    TEXT NOT NULL DEFAULT '{}'
);

-- Índice, não UNIQUE(form_id, position): constraint única em posição
-- transforma qualquer reordenação num quebra-cabeça de updates temporários.
-- A ordenação é reescrita inteira a cada save.
CREATE INDEX idx_form_fields_form_id_position ON form_fields (form_id, position);

CREATE TABLE submissions (
    id           TEXT PRIMARY KEY,
    form_id      TEXT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    submitted_at TEXT NOT NULL
);

CREATE INDEX idx_submissions_form_id ON submissions (form_id, submitted_at);

CREATE TABLE submission_answers (
    id            TEXT PRIMARY KEY,
    submission_id TEXT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    field_id      TEXT NOT NULL,
    -- Cópia do label no momento do envio: se o admin renomear a pergunta
    -- depois, as respostas antigas continuam legíveis com o texto que o
    -- visitante realmente viu.
    field_label   TEXT NOT NULL,
    -- Sempre TEXT; o type do campo dá a interpretação na leitura.
    value         TEXT NOT NULL
);

CREATE INDEX idx_submission_answers_submission_id ON submission_answers (submission_id);

-- +goose Down
DROP TABLE submission_answers;
DROP TABLE submissions;
DROP TABLE form_fields;
DROP TABLE forms;
DROP TABLE users;
