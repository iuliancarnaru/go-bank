package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(uuid.UUID) error
	UpdateAccount(*Account) error
	GetAccountByID(uuid.UUID) (*Account, error)
	GetAccountByNumber(number int64) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

// var PG *PostgresStore

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "postgres://postgres:password@localhost/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY,
		first_name varchar(50),
		last_name varchar(50),
		encrypted_password varchar(100),
		number bigint,
		balance bigint,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `INSERT INTO accounts (
		id, first_name, last_name, encrypted_password, number, balance, created_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	)`

	_, err := s.db.Query(query, acc.ID, acc.FirstName, acc.LastName, acc.EncryptedPassword, acc.Number, acc.Balance, acc.CreatedAt)
	return err
}

func (s *PostgresStore) DeleteAccount(id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1`
	_, err := s.db.Query(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(acc *Account) error {
	return nil
}

func (s *PostgresStore) GetAccountByID(id uuid.UUID) (*Account, error) {
	query := `SELECT * FROM accounts WHERE id = $1`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %v not found", id)
}

func (s *PostgresStore) GetAccountByNumber(number int64) (*Account, error) {
	query := `SELECT * FROM accounts WHERE number = $1`
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %v not found", number)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	query := `SELECT * FROM accounts`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}

	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.EncryptedPassword,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)

	return account, err
}
