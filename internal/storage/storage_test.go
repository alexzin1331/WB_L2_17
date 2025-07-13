package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		m := NewManager()
		event := Event{
			UserID: 1,
			Date:   time.Now(),
			Title:  "Test Event",
			Text:   "Test Description",
		}

		err := m.CreateEvent(event)
		assert.NoError(t, err)
		assert.Len(t, m.events[1], 1)
		assert.Equal(t, event, m.events[1][0])
	})

	t.Run("UpdateEvent", func(t *testing.T) {
		m := NewManager()
		now := time.Now()
		initialEvent := Event{
			UserID: 1,
			Date:   now,
			Title:  "Initial Title",
			Text:   "Initial Text",
		}

		// Create event first
		m.CreateEvent(initialEvent)

		// Test successful update
		newTitle := "Updated Title"
		newText := "Updated Text"
		err := m.UpdateEvent(1, now, newTitle, newText)
		assert.NoError(t, err)
		assert.Equal(t, newTitle, m.events[1][0].Title)
		assert.Equal(t, newText, m.events[1][0].Text)

		// Test update non-existent event
		err = m.UpdateEvent(1, now.Add(24*time.Hour), "Title", "Text")
		assert.Error(t, err)
		assert.Equal(t, "event not found", err.Error())
	})

	t.Run("DeleteEvent", func(t *testing.T) {
		m := NewManager()
		now := time.Now()
		event := Event{
			UserID: 1,
			Date:   now,
			Title:  "Test Event",
			Text:   "Test Description",
		}

		// Create event first
		m.CreateEvent(event)

		// Test successful delete
		err := m.DeleteEvent(1, now)
		assert.NoError(t, err)
		assert.Len(t, m.events[1], 0)

		// Test delete non-existent event
		err = m.DeleteEvent(1, now)
		assert.Error(t, err)
		assert.Equal(t, "event not found", err.Error())
	})

	t.Run("GetEventsForRange", func(t *testing.T) {
		m := NewManager()
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		tomorrow := today.Add(24 * time.Hour)
		nextWeek := today.Add(7 * 24 * time.Hour)

		events := []Event{
			{UserID: 1, Date: today, Title: "Today 1", Text: "Event today"},
			{UserID: 1, Date: today, Title: "Today 2", Text: "Another event today"},
			{UserID: 1, Date: tomorrow, Title: "Tomorrow", Text: "Event tomorrow"},
			{UserID: 1, Date: nextWeek, Title: "Next Week", Text: "Event next week"},
			{UserID: 2, Date: today, Title: "User 2 Event", Text: "Different user"},
		}

		// Create all events
		for _, e := range events {
			m.CreateEvent(e)
		}

		tests := []struct {
			name     string
			userID   int
			start    time.Time
			end      time.Time
			expected []Event
		}{
			{
				name:     "Single day",
				userID:   1,
				start:    today,
				end:      today,
				expected: []Event{events[0], events[1]},
			},
			{
				name:     "Two days",
				userID:   1,
				start:    today,
				end:      tomorrow,
				expected: []Event{events[0], events[1], events[2]},
			},
			{
				name:     "Week range",
				userID:   1,
				start:    today,
				end:      nextWeek,
				expected: []Event{events[0], events[1], events[2], events[3]},
			},
			{
				name:     "Different user",
				userID:   2,
				start:    today,
				end:      today,
				expected: []Event{events[4]},
			},
			{
				name:     "No events",
				userID:   3,
				start:    today,
				end:      today,
				expected: []Event{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := m.GetEventsForRange(tt.userID, tt.start, tt.end)
				assert.ElementsMatch(t, tt.expected, result)
			})
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		m := NewManager()
		now := time.Now()
		event := Event{
			UserID: 1,
			Date:   now,
			Title:  "Concurrent Event",
			Text:   "Test concurrent access",
		}

		// Test concurrent create
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				m.CreateEvent(event)
			}()
		}
		wg.Wait()

		assert.Len(t, m.events[1], 100)

		// Test concurrent read while writing
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = m.GetEventsForRange(1, now.Add(-time.Hour), now.Add(time.Hour))
			}
		}()

		// Concurrent deletes
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = m.DeleteEvent(1, now)
			}()
		}
		wg.Wait()

		assert.Len(t, m.events[1], 50)
	})
}
