package util

import (
	"bufio"
	"fmt"
	"os"
)

type QuestionPrompter interface {
	ReadString(string) string
	ApplyDefault(string, string) string
	RepeatIfEmpty(string) string
	RepeatUntilValid(string, func(string) (bool, error)) string
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
	return qp.RepeatUntilValid(question, func(response string) (bool, error) {
		valid := len(response) > 0
		if !valid {
			return false, fmt.Errorf("Please insert a value.")
		}
		return valid, nil
	})
}

//ask the same question to the user until the isValid() callback returns true
func (qp QuestionPrompt) RepeatUntilValid(question string, isValid func(string) (bool, error)) string {
	var res string
	for {
		res = qp.ReadString(question)
		isValid, err := isValid(res)
		if err != nil {
			fmt.Println(err.Error())
		}
		if isValid {
			break
		}
	}
	return res
}
