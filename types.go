package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/flynn/hid"
	"github.com/flynn/u2f/u2fhid"
	"github.com/flynn/u2f/u2ftoken"
)

type Registration struct {
	Challenge []byte
	App       []byte
	KeyHandle []byte
}

type Device struct {
	DeviceInfo *hid.DeviceInfo
	Device     *u2fhid.Device
	Token      *u2ftoken.Token
	Version    string
	Result     []byte
	Request    u2ftoken.AuthenticateRequest

	Registration
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
