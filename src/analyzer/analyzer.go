package main

type BasicTextAnalyzer interface {
	contains(text string, keyWord string) float64
}

type Analyzer struct{}

func (a Analyzer) contains(text string, keyWord string) float64 {
	//TODO implement me
	panic("implement me")
}

func (a Analyzer) analyze(topics map[string]struct{}, message string) ([]string, error) {
	//TODO() нужно будет возвращать подходящие топики и ошибку
	panic("implement me")
}
