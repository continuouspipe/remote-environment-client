package util

import (
	"fmt"
	"bufio"
	"os"
)

type QuestionPrompter interface {
	ReadString(string) string
	ApplyDefault(string, string) string
	RepeatIfEmpty(string) string
}

type QuestionPrompt struct{}

func NewQuestionPrompt() *QuestionPrompt {
	return &QuestionPrompt{}
}

func (qp QuestionPrompt) ReadString(q string) string {
	fmt.Print(q, " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func (qp QuestionPrompt) ApplyDefault(question string, predef string) string {
	res := qp.ReadString(question)

	if res == "" && predef != "" {
		return predef
	}
	return res
}

func (qp QuestionPrompt) RepeatIfEmpty(question string) string {
	var res string
	for {
		res = qp.ReadString(question)
		if len(res) > 0 {
			break
		}
	}
	return res
}
