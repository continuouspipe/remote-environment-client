package git

import (
	"os"
	"bufio"
)

type Ignore struct {
	file string
	list []string
}

func NewIgnore() (*Ignore, error) {
	ignore := &Ignore{}
	ignore.file = ".gitignore"
	return ignore, nil
}

func (i *Ignore) loadFromIgnoreFile() error {
	file, err := os.OpenFile(i.file, os.O_RDWR|os.O_CREATE, 0664)
	defer file.Close()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i.list = append(i.list, scanner.Text())
	}
	return nil
}

func (i *Ignore) AddToIgnore(fileNames ...string) (bool, error) {
	err := i.loadFromIgnoreFile()
	if err != nil {
		return false, err
	}

	file, err := os.OpenFile(i.file, os.O_APPEND|os.O_WRONLY, 0664)
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
	for _, val := range i.list {
		if val == s {
			return true
		}
	}
	return false
}
