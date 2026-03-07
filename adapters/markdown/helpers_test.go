package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHeaderPath(t *testing.T) {
	current := []string{"H1"}
	level := 3
	title := "H3"

	got := updateHeaderPath(current, level, title)
	want := []string{"H1", "H3"}

	assert.Equal(t, want, got)
}
