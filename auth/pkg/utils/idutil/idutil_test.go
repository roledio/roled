package idutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		id1 := NewID()
		id2 := NewID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2, "IDs should be unique")
	})

	t.Run("generates IDs with expected length", func(t *testing.T) {
		id := NewID()
		// shortuuid v7 encodes 128-bit UUID to ~22 characters
		assert.Equal(t, 22, len(id), "ID should be 22 characters")
	})

	t.Run("uses DefaultChars character set", func(t *testing.T) {
		id := NewID()
		for _, char := range id {
			assert.Contains(t, DefaultChars, string(char),
				"ID should only contain characters from DefaultChars")
		}
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		id := NewID()
		// shortuuid DefaultChars excludes: 0, O, 1, I, l
		assert.NotContains(t, id, "0", "Should not contain '0'")
		assert.NotContains(t, id, "O", "Should not contain 'O'")
		assert.NotContains(t, id, "1", "Should not contain '1'")
		assert.NotContains(t, id, "I", "Should not contain 'I'")
		assert.NotContains(t, id, "l", "Should not contain 'l'")
	})

	t.Run("generates multiple unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		const iterations = 1000

		for i := 0; i < iterations; i++ {
			id := NewID()
			assert.False(t, ids[id], "ID should be unique: %s", id)
			ids[id] = true
		}
		assert.Equal(t, iterations, len(ids), "All IDs should be unique")
	})
}

func TestULID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		id1 := ULID()
		id2 := ULID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2, "IDs should be unique")
	})

	t.Run("generates IDs with expected length", func(t *testing.T) {
		id := ULID()
		// ULID is 26 characters (Crockford's Base32)
		assert.Equal(t, 26, len(id), "ULID should be 26 characters")
	})

	t.Run("uses Crockford's Base32 character set", func(t *testing.T) {
		id := ULID()
		// Crockford's Base32: 0-9, A-Z (excluding I, L, O, U)
		validPattern := regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
		assert.True(t, validPattern.MatchString(id),
			"ULID should match Crockford's Base32 pattern")
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		id := ULID()
		// Crockford's Base32 excludes: I, L, O, U
		assert.NotContains(t, id, "I", "Should not contain 'I'")
		assert.NotContains(t, id, "L", "Should not contain 'L'")
		assert.NotContains(t, id, "O", "Should not contain 'O'")
		assert.NotContains(t, id, "U", "Should not contain 'U'")
	})

	t.Run("generates multiple unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		const iterations = 1000

		for i := 0; i < iterations; i++ {
			id := ULID()
			assert.False(t, ids[id], "ULID should be unique: %s", id)
			ids[id] = true
		}
		assert.Equal(t, iterations, len(ids), "All ULIDs should be unique")
	})

	t.Run("monotonic within same millisecond", func(t *testing.T) {
		// Generate ULIDs rapidly; they should be monotonic due to locked monotonic reader
		var prevID string
		for i := 0; i < 100; i++ {
			id := ULID()
			if prevID != "" {
				// ULIDs should be strictly increasing when generated rapidly
				assert.GreaterOrEqual(t, id, prevID,
					"ULIDs should be monotonic: %s >= %s", id, prevID)
			}
			prevID = id
		}
	})
}

func TestNanoID(t *testing.T) {
	t.Run("generates unique IDs with default size", func(t *testing.T) {
		id1 := NanoID()
		id2 := NanoID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2, "IDs should be unique")
		assert.Equal(t, NanoIDDefaultSize, len(id1), "Default size should be used")
	})

	t.Run("generates IDs with custom size", func(t *testing.T) {
		customSize := 10
		id := NanoID(customSize)

		assert.NotEmpty(t, id)
		assert.Equal(t, customSize, len(id), "Custom size should be used")
	})

	t.Run("uses DefaultChars character set", func(t *testing.T) {
		id := NanoID()
		for _, char := range id {
			assert.Contains(t, DefaultChars, string(char),
				"ID should only contain characters from DefaultChars")
		}
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		id := NanoID()
		// DefaultChars excludes: 0, O, 1, I, l
		assert.NotContains(t, id, "0", "Should not contain '0'")
		assert.NotContains(t, id, "O", "Should not contain 'O'")
		assert.NotContains(t, id, "1", "Should not contain '1'")
		assert.NotContains(t, id, "I", "Should not contain 'I'")
		assert.NotContains(t, id, "l", "Should not contain 'l'")
	})

	t.Run("generates multiple unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		const iterations = 1000

		for i := 0; i < iterations; i++ {
			id := NanoID()
			assert.False(t, ids[id], "ID should be unique: %s", id)
			ids[id] = true
		}
		assert.Equal(t, iterations, len(ids), "All IDs should be unique")
	})

	t.Run("handles zero size gracefully", func(t *testing.T) {
		id := NanoID(0)
		assert.NotEmpty(t, id, "Should generate ID even with zero size")
		assert.Equal(t, NanoIDDefaultSize, len(id), "Should use default size for zero")
	})

	t.Run("handles negative size gracefully", func(t *testing.T) {
		id := NanoID(-5)
		assert.NotEmpty(t, id, "Should generate ID even with negative size")
		assert.Equal(t, NanoIDDefaultSize, len(id), "Should use default size for negative")
	})
}

func TestConstants(t *testing.T) {
	t.Run("DefaultChars is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, DefaultChars)
	})

	t.Run("NanoIDDefaultSize is positive", func(t *testing.T) {
		assert.Greater(t, NanoIDDefaultSize, 0)
	})
}

// Benchmark tests
func BenchmarkNewID(b *testing.B) {
	for b.Loop() {
		_ = NewID()
	}
}

func BenchmarkNanoID(b *testing.B) {
	for b.Loop() {
		_ = NanoID()
	}
}

func BenchmarkULID(b *testing.B) {
	for b.Loop() {
		_ = ULID()
	}
}
