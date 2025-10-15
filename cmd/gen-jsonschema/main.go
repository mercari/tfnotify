package main

import (
	"fmt"
	"log"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := jsonschema.Write(&config.Config{}, "json-schema/tfnotify.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	return nil
}
