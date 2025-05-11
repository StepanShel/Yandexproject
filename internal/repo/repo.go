package repo

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Repo struct {
	db *sql.DB
}

func NewRepository() (*Repo, error) {
	db, err := sql.Open("sqlite3", "TODO")
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Repo{db: db}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        )
    `)

	if err != nil {
		return err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS expressions (
            id TEXT PRIMARY KEY,
            username TEXT NOT NULL,
            expression TEXT NOT NULL,
            result INTEGER,
            status TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(username) REFERENCES users(username)
        )
    `)

	return err
}

// Users methods
// ------------------------------------------------------------------------//

func (r *Repo) InsertUser(user User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	_, err = r.db.Exec(
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username, string(hashedPassword))
	return err
}

func (r *Repo) GetUser(username, password string) (*User, error) {

	var user User
	var hashedPassword string

	err := r.db.QueryRow(
		"SELECT username, password FROM users WHERE username = ?",
		username,
	).Scan(&user.Username, &hashedPassword)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repo) Authenticate(username, password string) (bool, error) {
	var hashedPassword string
	err := r.db.QueryRow(
		"SELECT password FROM users WHERE username = ?",
		username,
	).Scan(&hashedPassword)

	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, nil
}

// ------------------------------------------------------------------------//

// Expressions methods
// ------------------------------------------------------------------------//

func (r *Repo) CreateExpression(expr *Expression) error {
	expr.ID = uuid.New()
	_, err := r.db.Exec(
		"INSERT INTO expressions (id, username, expression, status) VALUES ($1, $2, $3, $4)",
		expr.ID.String(), expr.Username, expr.Expression, expr.Status)
	return err
}

func (r *Repo) UpdateExpressionResult(id uuid.UUID, result int, status string) error {
	_, err := r.db.Exec(
		"UPDATE expressions SET result = ?, status = ? WHERE id = ?",
		result, status, id.String())
	return err
}

func (r *Repo) GetExpressions(username string) ([]Expression, error) {
	rows, err := r.db.Query(
		"SELECT id, expression, result, status, created_at FROM expressions WHERE username = ?",
		username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []Expression
	for rows.Next() {
		var expr Expression
		var idStr string

		err := rows.Scan(&idStr, &expr.Expression, &expr.Result, &expr.Status, &expr.CreatedAt)
		if err != nil {
			return nil, err
		}

		expr.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}

		expr.Username = username
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (r *Repo) GetExpressionByID(id uuid.UUID) (*Expression, error) {
	var expr Expression
	var idStr, username string

	err := r.db.QueryRow(
		"SELECT id, username, expression, result, status, created_at FROM expressions WHERE id = ?",
		id.String()).Scan(&idStr, &username, &expr.Expression, &expr.Result, &expr.Status, &expr.CreatedAt)
	if err != nil {
		return nil, err
	}

	expr.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	return &expr, nil
}

//------------------------------------------------------------------------//
