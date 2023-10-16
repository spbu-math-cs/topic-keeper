package main

type Summarizer interface {
	summarize(text string) (string, error)
}

type MessageSummarizer struct {
	textLength int
}

func (m MessageSummarizer) summarize(text string) (string, error) {
	testRunes := []rune(text)

	length := len(testRunes)
	if length > m.textLength {
		length = m.textLength
	}

	return string(testRunes[:length]), nil
}
