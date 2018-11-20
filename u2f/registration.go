package u2f

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Registration struct {
	Challenge   []byte
	Application []byte
	KeyHandle   []byte
}

func (reg *Registration) Dump() error {
	text, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("\n%s\n", string(text))
	return nil
}

func (reg *Registration) WriteFile(filename string) error {
	text, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, text, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (reg *Registration) ReadFile(filename string) error {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(text, reg)
	if err != nil {
		return err
	}
	return nil
}
