package main

type Summarizer interface {
	Summarize(text string) (string, error)
}

type MessageSummarizer struct {
	textLength int
}

func (m MessageSummarizer) Summarize(text string) (string, error) {
	length := min(m.textLength, len(text))
	return text[:length-1], nil
}
