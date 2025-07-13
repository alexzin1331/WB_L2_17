package storage

import (
	"errors"
	"sync"
	"time"
)

// Event represents a calendar event with user ID, date, title and description
type Event struct {
	UserID int       // unique identifier for the user
	Date   time.Time // date of the event in UTC
	Title  string    // short title of the event
	Text   string    // detailed description of the event
}

// Manager provides thread-safe storage for calendar events
type Manager struct {
	mu     sync.RWMutex    // mutex for concurrent access protection
	events map[int][]Event // map of user IDs to their events
}

// NewManager creates a new initialized Manager instance
func NewManager() *Manager {
	return &Manager{events: make(map[int][]Event)}
}

// CreateEvent adds a new event to the storage
func (c *Manager) CreateEvent(e Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events[e.UserID] = append(c.events[e.UserID], e)
	return nil
}

// UpdateEvent modifies an existing event's title and text
func (c *Manager) UpdateEvent(userID int, date time.Time, title, newText string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// search for event with matching date
	for i, e := range c.events[userID] {
		if e.Date.Equal(date) {
			// update found event
			c.events[userID][i].Text = newText
			c.events[userID][i].Title = title
			return nil
		}
	}
	return errors.New("event not found")
}

// DeleteEvent removes an event by user ID and date
func (c *Manager) DeleteEvent(userID int, date time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	events := c.events[userID]
	for i, e := range events {
		if e.Date.Equal(date) {
			// remove event by slicing the array
			c.events[userID] = append(events[:i], events[i+1:]...)
			return nil
		}
	}
	return errors.New("event not found")
}

// GetEventsForRange returns events within specified date range
func (c *Manager) GetEventsForRange(userID int, start, end time.Time) []Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []Event
	for _, e := range c.events[userID] {
		// check if event falls within the range
		if !e.Date.Before(start) && !e.Date.After(end) {
			result = append(result, e)
		}
	}
	return result
}
