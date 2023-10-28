package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"io/ioutil"
	"net/http"
)

type AnalyzerRequest struct {
	Text   string   `json:"text"`
	Topics []string `json:"topics"`
}

type AnalyzerReturn struct {
	Summary string   `json:"summary"`
	Topics  []string `json:"topics"`
}

var (
	analyzer   BasicTextAnalyzer
	summarizer Summarizer
)

const (
	summaryLength = 100
)

func main() {
	analyzer = &Analyzer{}
	summarizer = &MessageSummarizer{textLength: summaryLength}

	router := gin.Default()
	router.POST("/analyze", analyze)
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func setAnswer(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"message": message,
	})
}

func analyze(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var request AnalyzerRequest

	err = json.Unmarshal(body, &request)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	topics, err := analyzer.analyze(request.Topics, request.Text)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "analyzer error")
		return
	}

	summary, err := summarizer.summarize(request.Text)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "summarizer error")
		return
	}

	c.JSON(http.StatusOK, AnalyzerReturn{Topics: topics, Summary: summary})

}
