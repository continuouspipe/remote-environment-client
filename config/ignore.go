package config

import (
	"bufio"
	"os"
)

const GitIgnore = ".gitignore"

type Ignore struct {
	File string
	List []string
}

func NewIgnore() *Ignore {
	return &Ignore{}
}

func (i *Ignore) LoadFromIgnoreFile() error {
	file, err := os.OpenFile(i.File, os.O_RDWR|os.O_CREATE, 0664)
	defer file.Close()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i.List = append(i.List, scanner.Text())
	}
	return nil
}

func (i *Ignore) AddToIgnore(fileNames ...string) (bool, error) {
	err := i.LoadFromIgnoreFile()
	if err != nil {
		return false, err
	}

	file, err := os.OpenFile(i.File, os.O_APPEND|os.O_WRONLY, 0664)
	defer file.Close()
	if err != nil {
		return false, err
	}
	for _, f := range fileNames {
		if !i.AlreadyIgnored(f) {
			_, err := file.WriteString(f)
			if err != nil {
				return false, err
			}
			_, err = file.WriteString("\n")
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (i *Ignore) AlreadyIgnored(s string) bool {
	for _, val := range i.List {
		if val == s {
			return true
		}
	}
	return false
}
