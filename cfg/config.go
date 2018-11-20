package cfg

import (
	"encoding/json"
	"fmt"
	"log"

	cfg "github.com/davidwalter0/go-cfg"
)

var (
	Env       = &Environment{}
	Semaphore = new(int64)
)

type Environment struct {
	Filename  string `json:"filename" required:"true" doc:"filename persist state"`
	Debugging bool   `json:"debugging"`
	Tracing   bool   `json:"tracing"`
}

func Setup() error {
	if Env.Debugging {
		defer Env.Trace("Setup")()
	}
	return cfg.Env(Env)
}

func Usage() {
	cfg.Usage()
}

func (env *Environment) Trace(function string, args ...string) func() {
	if env.Tracing {
		text := function + "("
		for _, arg := range args {
			text += fmt.Sprintf("%s ", arg)
		}
		log.Println("Entering:", text)
	}
	return func() {
		if env.Tracing {
			text := function + ": "
			for _, arg := range args {
				text += fmt.Sprintf("%s ", arg)
			}
			log.Println("Leaving:", text)
		}
	}
}

func (env *Environment) String() string {
	text, err := json.Marshal(*env)
	if err != nil {
		return ""
	}
	return string(text)
}

func (env *Environment) Format() string {
	text, err := json.MarshalIndent(*env, "", "  ")
	if err != nil {
		return ""
	}
	return string(text)
}
