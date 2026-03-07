package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
)

func TestUpdateHeaderPath(t *testing.T) {
	t.Run("heading level jump", func(t *testing.T) {
		current := []string{"H1"}
		levels := []int{1}
		level := 3
		title := "H3"

		got, newLevels := updateHeaderPath(current, levels, level, title)
		want := []string{"H1", "H3"}

		assert.Equal(t, want, got)
		assert.Equal(t, []int{1, 3}, newLevels)
	})
}

func TestScanner_Scan(t *testing.T) {
	scanner := &MDScanner{}
	gm := goldmark.New(goldmark.WithExtensions(&wikilink.Extender{}))

	t.Run("splits by headings and extracts wikilinks", func(t *testing.T) {
		content := `## Section 1
   This is [[Note 1]]
   ## Section 2
   This is [[Note 2]]`

		source := []byte(content)
		reader := text.NewReader(source)
		docNode := gm.Parser().Parse(reader)

		got := scanner.Scan(docNode, source)

		assert.Equal(t, 2, len(got))
		assert.Equal(t, []string{"Section 1"}, got[0].HeaderPath)
		assert.Contains(t, got[0].Links, "Note 1")
		assert.Equal(t, []string{"Section 2"}, got[1].HeaderPath)
		assert.Contains(t, got[1].Links, "Note 2")
	})
	t.Run("nested headings", func(t *testing.T) {
		content := `# H1
## H2
### H3
## Another H2`

		source := []byte(content)
		reader := text.NewReader(source)
		docNode := gm.Parser().Parse(reader)

		got := scanner.Scan(docNode, source)

		assert.Equal(t, 4, len(got))
		assert.Equal(t, []string{"H1"}, got[0].HeaderPath)
		assert.Equal(t, []string{"H1", "H2"}, got[1].HeaderPath)
		assert.Equal(t, []string{"H1", "H2", "H3"}, got[2].HeaderPath)
		assert.Equal(t, []string{"H1", "Another H2"}, got[3].HeaderPath)
	})
}
