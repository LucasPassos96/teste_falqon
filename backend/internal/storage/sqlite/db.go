// Package sqlite abre o banco e aplica as migrations.
package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/pressly/goose/v3"

	// Driver em Go puro: sem cgo, sem toolchain C, sem Docker. É o que faz o
	// projeto compilar igual em Windows, macOS e Linux.
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open abre o banco com os pragmas necessários e confirma a conexão.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}

	// sql.Open não conecta de fato; sem isto, um caminho inválido só falharia
	// na primeira query, longe da causa.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("conectar em %s: %w", path, err)
	}

	return db, nil
}

// dsn monta a string de conexão com os pragmas no próprio DSN.
//
// Os pragmas precisam estar aqui e não num Exec avulso porque valem POR
// CONEXÃO, e o database/sql mantém um pool: um `PRAGMA` executado uma vez
// configuraria só a primeira conexão, e o bug apareceria de forma
// intermitente. No DSN, o driver aplica em cada conexão nova.
//
//   - foreign_keys: desligado por padrão no SQLite, por compatibilidade
//     retroativa. Sem ele, os ON DELETE CASCADE do schema são decorativos.
//   - journal_mode=WAL: leitores concorrentes com um escritor.
//   - busy_timeout: o escritor espera em vez de estourar "database is locked"
//     na hora.
func dsn(path string) string {
	return "file:" + url.PathEscape(filepath.ToSlash(path)) +
		"?_pragma=foreign_keys(1)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)"
}

// Migrate aplica as migrations pendentes. Elas vêm do embed.FS, então estão
// dentro do binário: não há como rodar o servidor sem os arquivos de migration
// por perto.
func Migrate(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("definir dialeto: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("aplicar migrations: %w", err)
	}
	return nil
}

// Version devolve a versão de migration atual do banco, para o log de boot.
func Version(db *sql.DB) (int64, error) {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return 0, err
	}
	return goose.GetDBVersion(db)
}
