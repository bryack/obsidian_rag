package tokenizer

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) ToSparseVector(text string) map[uint32]float32 {
	return map[uint32]float32{}
}
