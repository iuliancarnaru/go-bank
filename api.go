package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

type APIFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandleFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{
				Error: err.Error(),
			})
		}
	}
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleCreateAccount)).Methods(http.MethodPost)
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store)).Methods(http.MethodGet)
	router.HandleFunc("/account/{id}", makeHTTPHandleFunc(s.handleDeleteAccount)).Methods(http.MethodDelete)
	router.HandleFunc("/accounts", makeHTTPHandleFunc(s.handleGetAccounts)).Methods(http.MethodGet)
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer)).Methods(http.MethodPost)
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin)).Methods(http.MethodPost)
	router.HandleFunc("/logout", makeHTTPHandleFunc(s.handleLogout)).Methods(http.MethodPost)

	log.Println("Server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		return fmt.Errorf("invalid account id: %v", id)
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccRequest := CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&createAccRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	acc, err := NewAccount(createAccRequest.FirstName, createAccRequest.LastName, createAccRequest.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(acc); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, acc)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		return fmt.Errorf("invalid account id: %v", id)
	}

	// THIS IS ANOTHER METHOD INSTEAD USING withJWTAuth

	// user, err := s.fetchUser(w, r)
	// if err != nil {
	// 	return err
	// }

	// if user.ID != id {
	// 	return fmt.Errorf("permission denied")
	// }

	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("deleted account with id: %s", id)})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferRequest := TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(&transferRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferRequest)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	acc, err := s.store.GetAccountByNumber(int64(loginRequest.Number))
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(loginRequest.Password))
	if err != nil {
		return fmt.Errorf("unable to login, verify credentials")
	}

	token, err := generateJWT(acc)
	if err != nil {
		return err
	}

	tokenResp := TokenResponse{
		Token: token,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(5 * time.Minute),
		Secure:   true,
		HttpOnly: true,
	})

	return WriteJSON(w, http.StatusOK, tokenResp)
}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) error {
	// clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})

	return WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *APIServer) fetchUser(w http.ResponseWriter, r *http.Request) (*Account, error) {
	tokenStr := r.Header.Get("Authorization")

	token, err := validateJWT(tokenStr)
	if err != nil || !token.Valid {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	id, _ := uuid.Parse(claims["id"].(string))
	acc, err := s.store.GetAccountByID(id)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")

		token, err := validateJWT(tokenStr)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		id, _ := uuid.Parse(claims["id"].(string))
		acc, err := s.GetAccountByID(id)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "permission denied"})
			return
		}

		if acc.ID != id {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "permission denied"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	jtwSecret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenStr, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}

		return []byte(jtwSecret), nil
	})
}

func generateJWT(account *Account) (string, error) {
	jtwSecret := os.Getenv("JWT_SECRET")

	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"id":        account.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jtwSecret))
}
