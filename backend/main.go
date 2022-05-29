package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/go-chi/chi"
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidAddress = errors.New("invalid address")
)

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

var hexRegex *regexp.Regexp = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

type RegisterPayload struct {
	Address string `json:"address"`
}

func (p RegisterPayload) Validate() error {
	if !hexRegex.MatchString(p.Address) {
		return ErrInvalidAddress
	}
	return nil
}

func RegisterHandler(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p RegisterPayload
		if err := bindReqBody(r, &p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := p.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		u := User{
			Address: strings.ToLower(p.Address), // let's only store lower case
		}
		if err := storage.CreateIfNotExists(u); err != nil {
			switch errors.Is(err, ErrUserExists) {
			case true:
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
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

// ============================================================================

func bindReqBody(r *http.Request, obj any) error {
	return json.NewDecoder(r.Body).Decode(obj)
}

// ============================================================================

func run() error {
	// initialization of storage
	storage := NewMemStorage()

	// setup the endpoints
	r := chi.NewRouter()
	r.Post("/register", RegisterHandler(storage))
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
