package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "gobank"
	dbname   = "postgres"
)

type Storage interface {
	CreateAccount(*Account) error
	UpdateAccount(*Account) error
	DeleteAccount(int) error
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

// connect to postgres
func NewPostgresStore() (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
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

// create initial table
func (s *PostgresStore) createAccountTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS account (
		id serial PRIMARY KEY,
  	first_name varchar(45) NOT NULL,
  	last_name varchar(45) NOT NULL,
		number serial,
		balance integer,
		created_at timestamp
	);`

	_, err := s.db.Exec(sqlStatement)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	sqlStatement := `
		INSERT INTO account (first_name, last_name, number, balance, created_at)
		VALUES ($1, $2, $3, $4, $5);`

	_, err := s.db.Exec(sqlStatement, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	sqlStatement := `DELETE FROM account WHERE id = $1;`
	_, err := s.db.Query(sqlStatement, id)
	return err
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	sqlStatement := `SELECT * FROM account WHERE id=$1;`
	rows, err := s.db.Query(sqlStatement, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	sqlStatement := "SELECT * FROM account;"
	rows, err := s.db.Query(sqlStatement)
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
	account := new(Account)
	err := rows.Scan(
		&account.ID, &account.FirstName,
		&account.LastName, &account.Number,
		&account.Balance, &account.CreatedAt,
	)

	return account, err
}
