package util

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
)

type QuestionPrompter interface {
	ReadString(string) string
	ApplyDefault(string, string) string
	RepeatIfEmpty(string) string
	RepeatUntilValid(string, func(string) (bool, error)) string
	RepeatPasswordIfEmpty(string) string
	RepeatPasswordUntilValid(string, func(string) (bool, error)) string
	ReadPassword(string) string
}

type QuestionPrompt struct{}

func NewQuestionPrompt() *QuestionPrompt {
	return &QuestionPrompt{}
}

func (qp QuestionPrompt) ReadString(question string) string {
	fmt.Print(question, " ")
	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		res = ""
	}
	return strings.TrimRight(res, "\n")
}

func (qp QuestionPrompt) ReadPassword(q string) string {
	fmt.Print(q, " ")
	res, err := terminal.ReadPassword(0)
	if err != nil {
		res = []byte{}
	}
	return string(res)
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

func (qp QuestionPrompt) RepeatPasswordIfEmpty(question string) string {
	return qp.RepeatPasswordUntilValid(question, func(response string) (bool, error) {
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

//ask the same question to the user until the isValid() callback returns true
func (qp QuestionPrompt) RepeatPasswordUntilValid(question string, isValid func(string) (bool, error)) string {
	var res string
	for {
		res = qp.ReadPassword(question)
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
