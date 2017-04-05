package util

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
)

// QuestionPrompter prints a question in the stdout and handle the response by applying a default value or repeating the question
// until a specific condition is validated
type QuestionPrompter interface {
	ApplyDefault(string, string) string
	RepeatIfEmpty(string) string
	RepeatUntilValid(string, func(string) (bool, error)) string
	RepeatPasswordIfEmpty(string) string
	RepeatPasswordUntilValid(string, func(string) (bool, error)) string
}

// QuestionPrompt implements QuestionPrompter and allows us to request the user a question and handles the answer in different ways
type QuestionPrompt struct{}

// NewQuestionPrompt constructor for QuestionPrompt
func NewQuestionPrompt() *QuestionPrompt {
	return &QuestionPrompt{}
}

// ApplyDefault if the answer is empty it returns a default value
func (qp QuestionPrompt) ApplyDefault(question string, predef string) string {
	res := qp.readString(question)

	if res == "" && predef != "" {
		return predef
	}
	return res
}

// RepeatIfEmpty requires the user to answer the question. It repeats the question until he answers.
func (qp QuestionPrompt) RepeatIfEmpty(question string) string {
	return qp.RepeatUntilValid(question, func(response string) (bool, error) {
		valid := len(response) > 0
		if !valid {
			return false, fmt.Errorf("please insert a value")
		}
		return valid, nil
	})
}

// RepeatPasswordIfEmpty requires the user to answer the question by inserting an answer that wont be shown when is typed.
// It repeats the question until he answers.
func (qp QuestionPrompt) RepeatPasswordIfEmpty(question string) string {
	return qp.RepeatPasswordUntilValid(question, func(response string) (bool, error) {
		valid := len(response) > 0
		if !valid {
			return false, fmt.Errorf("please insert a value")
		}
		return valid, nil
	})
}

// RepeatUntilValid ask the same question to the user until the isValid() callback returns true
func (qp QuestionPrompt) RepeatUntilValid(question string, isValid func(string) (bool, error)) string {
	var res string
	for {
		res = qp.readString(question)
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

// RepeatPasswordUntilValid ask the same question to the user until the isValid() callback returns true
func (qp QuestionPrompt) RepeatPasswordUntilValid(question string, isValid func(string) (bool, error)) string {
	var res string
	for {
		res = qp.readPassword(question)
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

func (qp QuestionPrompt) readString(question string) string {
	fmt.Print(question, " ")
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		line = []byte{}
	}

	return strings.TrimRight(string(line), " \r\n")
}

func (qp QuestionPrompt) readPassword(q string) string {
	fmt.Print(q, " ")
	res, err := terminal.ReadPassword(0)
	if err != nil {
		res = []byte{}
	}
	return string(res)
}
