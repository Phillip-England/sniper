package sniper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// MouseSpot represents a saved X, Y coordinate.
type MouseSpot struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// MouseMemory manages the persistence of mouse locations.
type MouseMemory struct {
	Spots    map[string]MouseSpot `json:"spots"`
	FilePath string
	mu       sync.RWMutex
}

// NewMouseMemory creates the manager and loads existing spots.
func NewMouseMemory() *MouseMemory {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".sniper_spots.json")

	mm := &MouseMemory{
		Spots:    make(map[string]MouseSpot),
		FilePath: path,
	}
	mm.Load()
	return mm
}

// Load reads the JSON file from disk.
func (mm *MouseMemory) Load() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	data, err := os.ReadFile(mm.FilePath)
	if err != nil {
		// If file doesn't exist, start fresh
		return
	}

	json.Unmarshal(data, &mm.Spots)
}

// Save writes the current map to disk.
func (mm *MouseMemory) Save() {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	data, err := json.MarshalIndent(mm.Spots, "", "  ")
	if err != nil {
		fmt.Printf("Error saving mouse memory: %v\n", err)
		return
	}

	os.WriteFile(mm.FilePath, data, 0644)
}

// Set saves a coordinate with a name (normalized to lower case).
func (mm *MouseMemory) Set(name string, x, y int) {
	mm.mu.Lock()
	name = strings.ToLower(name)
	mm.Spots[name] = MouseSpot{X: x, Y: y}
	mm.mu.Unlock()
	mm.Save()
}

// Get retrieves a coordinate. Returns bool indicating existence.
func (mm *MouseMemory) Get(name string) (MouseSpot, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	name = strings.ToLower(name)
	val, ok := mm.Spots[name]
	return val, ok
}

// Delete removes a spot.
func (mm *MouseMemory) Delete(name string) {
	mm.mu.Lock()
	name = strings.ToLower(name)
	delete(mm.Spots, name)
	mm.mu.Unlock()
	mm.Save()
}
