package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	mux  *sync.RWMutex
	path string
}

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	ExpirationTime time.Time `json:"expiration_time"`
	Password       string    `json:"password"`
	Email          string    `json:"email"`
	RefreshToken   string    `json:"refresh_token"`
	Id             int       `json:"id"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{
		mux:  &sync.RWMutex{},
		path: path,
	}

	err := db.ensureDB()

	return &db, err
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	database, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	max := 0
	for key := range database.Chirps {
		if key > max {
			max = key
		}
	}
	chirp := Chirp{
		Body:     body,
		Id:       max + 1,
		AuthorId: authorId,
	}
	database.Chirps[chirp.Id] = chirp
	err = db.writeDB(database)
	if err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

// CreateUser creates a new user and saves it to disk
func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	database, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	max := 0
	for key, value := range database.Users {
		if key > max {
			max = key
		}
		if strings.EqualFold(value.Email, email) {
			return User{}, errors.New("a user with this email already exists")
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return User{}, err
	}
	user := User{
		Password: string(hashedPassword),
		Email:    email,
		Id:       max + 1,
	}
	database.Users[user.Id] = user
	err = db.writeDB(database)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (db *DB) LoginUser(email string, password string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	database, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user := User{}
	for _, value := range database.Users {
		if strings.EqualFold(value.Email, email) {
			user = value
		}
	}
	if len(user.Email) <= 0 {
		return User{}, errors.New("the user doesn't exist")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, err
	}
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return User{}, err
	}
	user.RefreshToken = hex.EncodeToString(b)
	user.ExpirationTime = time.Now().AddDate(0, 0, 60)
	database.Users[user.Id] = user
	db.writeDB(database)
	return user, nil
}

func (db *DB) RevokeRefreshToken(token string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	database, err := db.loadDB()
	if err != nil {
		return err
	}
	user := User{}
	for _, value := range database.Users {
		if strings.EqualFold(value.RefreshToken, token) {
			user = value
		}
	}
	if user.Id != 0 {
		user.RefreshToken = ""
		user.ExpirationTime = time.Now()
		database.Users[user.Id] = user
		db.writeDB(database)
	}
	return nil
}

func (db *DB) GetUserByRefreshToken(token string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	database, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user := User{}
	for _, value := range database.Users {
		if strings.EqualFold(value.RefreshToken, token) {
			user = value
		}
	}
	return user, nil
}

// UpdateUser user updates the given user and returns the updated user
func (db *DB) UpdateUser(id string, email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	database, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user := User{}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return User{}, err
	}
	for _, value := range database.Users {
		if value.Id == idInt {
			user = value
		}
	}
	if len(password) > 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return User{}, err
		}
		user.Password = string(hashedPassword)
	}
	if len(email) > 0 {
		user.Email = email
	}
	database.Users[user.Id] = user
	err = db.writeDB(database)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}
	chirps := []Chirp{}
	for _, value := range database.Chirps {
		chirps = append(chirps, value)
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})
	return chirps, nil
}

func (db *DB) GetSingleChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	database, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := database.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("doesn't exist")
	}
	return chirp, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	stat, err := os.Stat(db.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return db.initializeDB()
		}
		return err
	}

	if stat.Size() == 0 {
		return db.initializeDB()
	}

	return nil
}

func (db *DB) initializeDB() error {
	file, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Initialize with an empty structure
	dbStructure := DBStructure{
		Chirps: make(map[int]Chirp),
		Users:  map[int]User{},
	}

	// Convert the structure to JSON and write it to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(dbStructure); err != nil {
		return err
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	dbStructure := DBStructure{}

	bs, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	err = json.Unmarshal(bs, &dbStructure)
	if err != nil {
		return DBStructure{}, err
	}

	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
