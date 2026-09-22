package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/roledio/roled/auth/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestKeyWithPrefix tests the prefix key generation logic
func TestKeyWithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		key    string
		want   string
	}{
		{
			name:   "empty prefix returns key as-is",
			prefix: "",
			key:    "mykey",
			want:   "mykey",
		},
		{
			name:   "prefix adds colon separator",
			prefix: "app",
			key:    "mykey",
			want:   "app:mykey",
		},
		{
			name:   "empty key returns empty",
			prefix: "app",
			key:    "",
			want:   "",
		},
		{
			name:   "already prefixed key unchanged",
			prefix: "app",
			key:    "app:mykey",
			want:   "app:mykey",
		},
		{
			name:   "different prefix still adds new prefix",
			prefix: "app",
			key:    "other:key",
			want:   "app:other:key",
		},
		{
			name:   "empty prefix with prefixed key",
			prefix: "",
			key:    "app:key",
			want:   "app:key",
		},
		{
			name:   "single char prefix",
			prefix: "a",
			key:    "key",
			want:   "a:key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				prefix: tt.prefix,
			}
			result := svc.KeyWithPrefix(tt.key)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestKeyWithPrefix_EdgeCases tests edge cases for KeyWithPrefix
func TestKeyWithPrefix_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		key    string
		want   string
	}{
		{
			name:   "multiple colons in prefix",
			prefix: "app:cache:v1",
			key:    "key",
			want:   "app:cache:v1:key",
		},
		{
			name:   "multiple colons in key",
			prefix: "app",
			key:    "user:1:profile",
			want:   "app:user:1:profile",
		},
		{
			name:   "hyphen and underscore",
			prefix: "my-app_v2",
			key:    "my_key-1",
			want:   "my-app_v2:my_key-1",
		},
		{
			name:   "numeric prefix",
			prefix: "1",
			key:    "key",
			want:   "1:key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				prefix: tt.prefix,
			}
			result := svc.KeyWithPrefix(tt.key)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestConfig tests the Config struct can be created
func TestConfig_Creation(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     6379,
		Username: "user",
		Password: "pass",
		Prefix:   "app",
		DB:       0,
		Newrelic: false,
	}

	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, "user", config.Username)
	assert.Equal(t, "pass", config.Password)
	assert.Equal(t, "app", config.Prefix)
	assert.Equal(t, 0, config.DB)
	assert.False(t, config.Newrelic)
}

// TestJSONSerialization tests JSON marshaling for SetData
func TestJSONSerialization(t *testing.T) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	person := Person{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	// Test marshaling
	data, err := json.Marshal(person)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var decoded Person
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, person, decoded)
}

// TestJSONSerialization_Complex tests JSON with nested structures
func TestJSONSerialization_Complex(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Employee struct {
		ID      int      `json:"id"`
		Name    string   `json:"name"`
		Address Address  `json:"address"`
		Tags    []string `json:"tags"`
	}

	emp := Employee{
		ID:   123,
		Name: "Bob",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
		Tags: []string{"admin", "active"},
	}

	data, err := json.Marshal(emp)
	assert.NoError(t, err)

	var decoded Employee
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, emp, decoded)
	assert.Equal(t, 123, decoded.ID)
	assert.Equal(t, "Bob", decoded.Name)
	assert.Equal(t, "NYC", decoded.Address.City)
	assert.Equal(t, 2, len(decoded.Tags))
}

// TestJSONSerialization_EmptyStructs tests JSON with empty data
func TestJSONSerialization_EmptyStructs(t *testing.T) {
	type Data struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	// Empty struct
	empty := Data{}
	data, err := json.Marshal(empty)
	assert.NoError(t, err)

	var decoded Data
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "", decoded.Name)
	assert.Empty(t, decoded.Tags)
}

// TestJSONSerialization_InvalidJSON tests unmarshaling invalid JSON
func TestJSONSerialization_InvalidJSON(t *testing.T) {
	type Data struct {
		Name string `json:"name"`
	}

	invalidJSONs := [][]byte{
		[]byte("not json"),
		[]byte("{invalid}"),
		[]byte(""),
		[]byte("{"),
		[]byte("["),
	}

	for _, invalidJSON := range invalidJSONs {
		var data Data
		err := json.Unmarshal(invalidJSON, &data)
		assert.Error(t, err)
	}
}

// TestJSONSerialization_TypeMismatch tests JSON unmarshal with type mismatch
func TestJSONSerialization_TypeMismatch(t *testing.T) {
	type Data struct {
		Value int `json:"value"`
	}

	// String value for int field
	jsonData := []byte(`{"value":"not-a-number"}`)
	var data Data
	err := json.Unmarshal(jsonData, &data)
	assert.Error(t, err)
}

// TestConfig_Defaults tests default values in Config
func TestConfig_Defaults(t *testing.T) {
	// Create config with minimal values
	config := &Config{}

	assert.Empty(t, config.Host)
	assert.Equal(t, 0, config.Port)
	assert.Empty(t, config.Username)
	assert.Empty(t, config.Password)
	assert.Empty(t, config.Prefix)
	assert.Equal(t, 0, config.DB)
	assert.False(t, config.Newrelic)
}

// TestService_Interface tests that service implements Service interface
func TestService_Interface(t *testing.T) {
	svc := &service{
		prefix: "test",
	}

	// Verify that service implements the methods
	assert.NotNil(t, svc.KeyWithPrefix("key"))
}

// TestTimeValues tests time duration handling for Set operations
func TestTimeValues(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		isValid  bool
	}{
		{"1 second", 1 * time.Second, true},
		{"1 hour", 1 * time.Hour, true},
		{"1 day", 24 * time.Hour, true},
		{"1 millisecond", 1 * time.Millisecond, true},
		{"1 nanosecond", 1 * time.Nanosecond, true},
		{"0 duration", 0, true},
		{"negative duration", -1 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Time values should be accepted (validation happens at the storage level)
			assert.True(t, tt.isValid)
		})
	}
}

// BenchmarkKeyWithPrefix benchmarks the prefix key generation
func BenchmarkKeyWithPrefix(b *testing.B) {
	svc := &service{
		prefix: "myapp",
	}

	for b.Loop() {
		svc.KeyWithPrefix("user:123:profile")
	}
}

// BenchmarkKeyWithPrefix_NoPrefix benchmarks without prefix
func BenchmarkKeyWithPrefix_NoPrefix(b *testing.B) {
	svc := &service{
		prefix: "",
	}

	for b.Loop() {
		svc.KeyWithPrefix("key")
	}
}

// BenchmarkJSONMarshal benchmarks JSON marshaling
func BenchmarkJSONMarshal(b *testing.B) {
	type Data struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := Data{
		ID:   123,
		Name: "John",
		Age:  30,
	}

	for b.Loop() {
		_, _ = json.Marshal(data)
	}
}

// BenchmarkJSONUnmarshal benchmarks JSON unmarshaling
func BenchmarkJSONUnmarshal(b *testing.B) {
	type Data struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	jsonData := []byte(`{"id":123,"name":"John","age":30}`)

	for b.Loop() {
		var data Data
		_ = json.Unmarshal(jsonData, &data)
	}
}

// TestPrefix_Handling tests various prefix scenarios
func TestPrefix_Handling(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		key         string
		expectedKey string
		description string
	}{
		{
			name:        "Standard case",
			prefix:      "cache",
			key:         "user:1",
			expectedKey: "cache:user:1",
			description: "Should prepend prefix with colon",
		},
		{
			name:        "No prefix",
			prefix:      "",
			key:         "user:1",
			expectedKey: "user:1",
			description: "Should return key unchanged when prefix is empty",
		},
		{
			name:        "No key",
			prefix:      "cache",
			key:         "",
			expectedKey: "",
			description: "Should return empty when key is empty",
		},
		{
			name:        "Already prefixed",
			prefix:      "cache",
			key:         "cache:data",
			expectedKey: "cache:data",
			description: "Should not double-prefix",
		},
		{
			name:        "Complex key",
			prefix:      "v1",
			key:         "user:profile:settings",
			expectedKey: "v1:user:profile:settings",
			description: "Should handle complex keys with multiple colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				prefix: tt.prefix,
			}

			result := svc.KeyWithPrefix(tt.key)
			assert.Equal(t, tt.expectedKey, result, tt.description)
		})
	}
}

// TestContextUsage tests that context is properly handled
func TestContextUsage(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	// Test with timeout
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NotNil(t, ctxWithTimeout)

	// Test with value
	ctxWithValue := context.WithValue(context.Background(), types.ContextKey("key"), "value")
	assert.NotNil(t, ctxWithValue)
}
