package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateIDF(t *testing.T) {
	stats := NewBM25Stats(1.5, 0.75)
	stats.DocsNumber = 1000
	stats.DocFrequency["rare"] = 5
	stats.DocFrequency["common"] = 500
	idf1 := stats.CalculateIDF("rare")
	idf2 := stats.CalculateIDF("common")

	assert.True(t, idf1 > idf2)
}

func TestCalculateTF(t *testing.T) {
	t.Run("term frequency saturation", func(t *testing.T) {
		stats := NewBM25Stats(1.5, 0.75)
		stats.AverageLength = 100
		docLen := 100

		tf1 := stats.CalculateTF(1, docLen)
		tf2 := stats.CalculateTF(10, docLen)

		assert.True(t, tf2 > tf1)
		assert.True(t, tf2 < tf1*10)
		maxTF := stats.k1 + 1 // теоретический максимум при freq -> ∞
		assert.True(t, tf1 > 0 && tf1 <= maxTF)
		assert.True(t, tf2 > 0 && tf2 <= maxTF)
	})

	t.Run("length penalty", func(t *testing.T) {
		stats := NewBM25Stats(1.5, 0.75)
		stats.AverageLength = 100
		tf1 := stats.CalculateTF(2, 50)
		tf2 := stats.CalculateTF(2, 200)

		assert.True(t, tf1 > tf2)
	})
}

func TestCalculateScore(t *testing.T) {
	stats := NewBM25Stats(1.5, 0.75)
	stats.DocsNumber = 1000
	stats.AverageLength = 100
	stats.DocFrequency["docker"] = 100
	stats.DocFrequency["go"] = 50
	queryTerms := map[string]int{"docker": 1, "go": 1}
	docTerm := map[string]int{"docker": 1}

	docLen := 50
	score := stats.CalculateScore(queryTerms, docTerm, docLen)

	assert.Equal(t, stats.CalculateTF(1, docLen)*stats.CalculateIDF("docker"), score)

	docBoth := map[string]int{"docker": 1, "go": 1}
	scoreBoth := stats.CalculateScore(queryTerms, docBoth, docLen)
	assert.True(t, score < scoreBoth, "partial match should score lower than full match")
}
