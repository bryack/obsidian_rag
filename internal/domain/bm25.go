package domain

import "math"

// BM25Stats хранит статистику корпуса документов для расчёта BM25.
// BM25 (Best Matching 25) — алгоритм ранжирования, который оценивает релевантность
// документа запросу на основе частоты термов с учётом их редкости в корпусе.
// Использует двухпроходный подход: Pass 1 собирает статистику по всем документам,
// Pass 2 индексирует с BM25-весами только изменённые документы.
type BM25Stats struct {
	// DocsNumber — общее количество документов (чанков) в корпусе.
	// Используется для расчёта IDF (Inverse Document Frequency).
	DocsNumber int `json:"docs_number"`

	// AverageLength — средняя длина документа в термах (словах после токенизации).
	// Используется для нормализации длины документа в формуле TF.
	AverageLength float64 `json:"average_length"`

	// DocFrequency — карта частоты документов (document frequency).
	// Ключ: терм (строка), Значение: сколько документов содержат этот терм.
	// Используется для определения редкости терма в корпусе.
	DocFrequency map[string]int `json:"doc_frequency"`

	// K1 — параметр насыщения частоты терма (term frequency saturation).
	// Контролирует, насколько сильно влияет повторение терма в документе.
	// При K1=0 влияние линейное, при K1-∞ почти постоянное (сатурация).
	// Обычные значения: 1.2–2.0, по умолчанию 1.5.
	K1 float64 `json:"k1"`

	// B — параметр нормализации длины документа.
	// Контролирует штраф за длинные документы.
	// При B=0 длина не учитывается, при B=1 полная нормализация.
	// Обычное значение: 0.75.
	B float64 `json:"b"`
}

// NewBM25Stats создаёт новый экземпляр BM25Stats с заданными параметрами.
// Если k1 или b равны 0, используются значения по умолчанию (1.5 и 0.75 соответственно).
// Инициализирует пустую карту DocFrequency для последующего заполнения в Pass 1.
func NewBM25Stats(k1, b float64) *BM25Stats {
	if k1 == 0 {
		k1 = 1.5
	}
	if b == 0 {
		b = 0.75
	}
	return &BM25Stats{
		DocFrequency: map[string]int{},
		K1:           k1,
		B:            b,
	}
}

// CalculateIDF вычисляет Inverse Document Frequency (обратную частоту документа)
// для заданного терма по формуле: log((N - df + 0.5) / (df + 0.5)),
// где N — общее число документов, df — частота документа для терма.
// Если терм отсутствует в DocFrequency, df считается равным 0.
// Результат ограничен снизу нулём (через math.Max), чтобы избежать
// отрицательных значений для очень частых термов.
func (s *BM25Stats) CalculateIDF(term string) float64 {
	freq := s.DocFrequency[term]

	numerator := float64(s.DocsNumber) - float64(freq) + 0.5
	denominator := float64(freq) + 0.5

	return math.Max(0, math.Log(numerator/denominator))
}

// CalculateTF вычисляет Term Frequency (нормализованную частоту терма)
// по формуле: (freq * (k1 + 1)) / (freq + k1 * (1 - b + b * docLen/avgdl)),
// где freq — частота терма в документе, docLen — длина документа,
// avgdl — средняя длина документа в корпусе.
// Реализует насыщение: первые вхождения терма дают больший вес,
// последующие — всё меньше ( diminishing returns).
// Учитывает длину документа: длинные документы штрафуются при b > 0.
// Возвращает 0, если AverageLength равен 0 (защита от деления на ноль).
func (s *BM25Stats) CalculateTF(freq int, docLen int) float64 {
	if s.AverageLength == 0 {
		return 0
	}
	numerator := float64(freq) * (s.K1 + 1)
	denominator := float64(freq) + s.K1*(1-s.B+s.B*float64(docLen)/s.AverageLength)
	return numerator / denominator
}

// CalculateScore вычисляет общий BM25-скор документа для заданного запроса.
// Суммирует взвешенные совпадения термов: score = sum(TF(term) * IDF(term))
// для всех термов запроса, которые присутствуют в документе.
// Параметры:
//   - queryTerms: термы запроса с их частотами (map[term]frequency)
//   - docTerms: термы документа с их частотами (map[term]frequency)
//   - docLen: длина документа в термах
//
// Чем выше возвращаемое значение, тем более релевантен документ запросу.
func (s *BM25Stats) CalculateScore(queryTerms, docTerms map[string]int, docLen int) float64 {
	score := 0.
	for term := range queryTerms {
		freq, ok := docTerms[term]
		if !ok {
			continue
		}
		TF := s.CalculateTF(freq, docLen)
		IDF := s.CalculateIDF(term)
		score += TF * IDF
	}
	return score
}
