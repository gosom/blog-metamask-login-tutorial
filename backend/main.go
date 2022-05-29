package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
)

var ErrUserExists = errors.New("user already exists")

type User struct {
	Address string
}

type MemStorage struct {
	lock  sync.Mutex
	users map[string]User
}

func (m *MemStorage) CreateIfNotExists(u User) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, exists := m.users[u.Address]; exists {
		return ErrUserExists
	}
	m.users[u.Address] = u
	return nil
}

func NewMemStorage() *MemStorage {
	ans := MemStorage{
		users: make(map[string]User),
	}
	return &ans
}

// ============================================================================

func RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func UserNonceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func SigninHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func WelcomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func run() error {
	// setup the endpoints
	r := chi.NewRouter()
	r.Post("/register", RegisterHandler())
	r.Get("/users/{address:^0x[a-fA-F0-9]{40}$}/nonce", UserNonceHandler())
	r.Post("/signin", SigninHandler())
	r.Get("/welcome", WelcomeHandler())

	// start the server on port 8001
	err := http.ListenAndServe("localhost:8001", r)
	return err
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err.Error())
	}
}
