package internal

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestDays_Next(t *testing.T) {
	day := Day{
		Sequence: []string{"a", "b", "random", "c", "random"},
		Random:   []string{"rand1", "rand2", "rand3"},
	}

	assert.EqualValues(t, "a", day.Next(time.Second))
	assert.EqualValues(t, "b", day.Next(time.Second))
	assert.True(t, strings.HasPrefix(day.Next(time.Second), "rand"))
	assert.EqualValues(t, "c", day.Next(time.Second))

	for i := 0; i < 100; i++ {
		assert.True(t, strings.HasPrefix(day.Next(time.Second), "rand"))
	}
}
