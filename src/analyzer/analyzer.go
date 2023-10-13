package main

type Analyzer struct {
	summarizer Summarizer
}

func (a *Analyzer) analyze(topics map[string]struct{}, message string) (map[string]struct{}, string, error) {
	//TODO() нужно будет возвращать подходящие топики, summary сообщения и ошибку
	panic("implement me")
}
