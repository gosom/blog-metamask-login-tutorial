package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

var (
	ErrUserNotExists  = errors.New("user does not exist")
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidAddress = errors.New("invalid address")
)

type User struct {
	Address string
	Nonce   string
}

type MemStorage struct {
	lock  sync.RWMutex
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

func (m *MemStorage) Get(address string) (User, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	u, exists := m.users[address]
	if !exists {
		return u, ErrUserNotExists
	}
	return u, nil
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
		nonce, err := GetNonce()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		u := User{
			Address: strings.ToLower(p.Address), // let's only store lower case
			Nonce:   nonce,
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

func UserNonceHandler(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := chi.URLParam(r, "address")
		if !hexRegex.MatchString(address) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := storage.Get(strings.ToLower(address))
		if err != nil {
			switch errors.Is(err, ErrUserNotExists) {
			case true:
				w.WriteHeader(http.StatusNotFound)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		resp := struct {
			Nonce string
		}{
			Nonce: user.Nonce,
		}
		renderJson(r, w, http.StatusOK, resp)
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

var (
	max  *big.Int
	once sync.Once
)

func GetNonce() (string, error) {
	once.Do(func() {
		max = new(big.Int)
		max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))
	})
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return n.Text(10), nil
}

func bindReqBody(r *http.Request, obj any) error {
	return json.NewDecoder(r.Body).Decode(obj)
}

func renderJson(r *http.Request, w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8 ")
	var body []byte
	if res != nil {
		var err error
		body, err = json.Marshal(res)
		if err != nil { // TODO handle me better
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	w.WriteHeader(statusCode)
	if len(body) > 0 {
		w.Write(body)
	}
}

// ============================================================================

func run() error {
	// initialization of storage
	storage := NewMemStorage()

	// setup the endpoints
	r := chi.NewRouter()

	//  Just allow all for the reference implementation
	r.Use(cors.AllowAll().Handler)

	r.Post("/register", RegisterHandler(storage))
	r.Get("/users/{address:^0x[a-fA-F0-9]{40}$}/nonce", UserNonceHandler(storage))
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
