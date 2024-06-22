package database

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
)

type DB struct {
	mux  *sync.RWMutex
	path string
}

type Chirp struct {
	Body string
	Id   int
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
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
func (db *DB) CreateChirp(body string) (Chirp, error) {
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
		Body: body,
		Id:   max + 1,
	}
	database.Chirps[chirp.Id] = chirp
	err = db.writeDB(database)
	if err != nil {
		return Chirp{}, err
	}
	return chirp, nil
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

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if _, err := os.Stat(db.path); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(db.path)
		if err != nil {
			return err
		}
		// Initialize with an empty structure
		dbStructure := DBStructure{
			Chirps: make(map[int]Chirp),
		}

		// Convert the structure to JSON and write it to the file
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(dbStructure); err != nil {
			file.Close()
			return err
		}

		file.Close()
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
