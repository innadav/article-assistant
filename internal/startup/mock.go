package startup

import (
	"context"
	"fmt"
	"log"
)

// MockLoader implements LoadOnStartupData for testing
type MockLoader struct {
	ShouldFail bool
	LoadCount  int
	Data       []string
}

// NewMockLoader creates a new MockLoader
func NewMockLoader() *MockLoader {
	return &MockLoader{
		ShouldFail: false,
		LoadCount:  0,
		Data:       make([]string, 0),
	}
}

// LoadData simulates loading data for testing
func (ml *MockLoader) LoadData(ctx context.Context, dataSource string) error {
	ml.LoadCount++
	ml.Data = append(ml.Data, dataSource)

	log.Printf("🔧 Mock loading data from: %s (count: %d)", dataSource, ml.LoadCount)

	if ml.ShouldFail {
		return fmt.Errorf("mock loader configured to fail")
	}

	log.Printf("✅ Mock data loading successful")
	return nil
}

// Reset resets the mock loader state
func (ml *MockLoader) Reset() {
	ml.ShouldFail = false
	ml.LoadCount = 0
	ml.Data = ml.Data[:0]
}
