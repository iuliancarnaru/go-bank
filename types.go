package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type TransferRequest struct {
	ToAccount uuid.UUID `json:"to_account"`
	Amount    int       `json:"amount"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type Account struct {
	ID                uuid.UUID `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	EncryptedPassword string    `json:"-"`
	Number            int64     `json:"number"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName, password string) (*Account, error) {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:                uuid.New(), // we can use postgres auto incremented id
		FirstName:         firstName,
		LastName:          lastName,
		EncryptedPassword: string(encryptedPassword),
		Number:            int64(rand.Intn(1000000)),
		Balance:           0,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
