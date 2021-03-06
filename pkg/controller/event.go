package controller

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v43/github"
)

type payloadReader struct{}

func newPayloadReader() PayloadReader {
	return &payloadReader{}
}

func (reader *payloadReader) Read(p string, payload *github.DiscussionEvent) error {
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("open a payload file %s: %w", p, err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(payload); err != nil {
		return fmt.Errorf("parse payload as JSON: %w", err)
	}
	return nil
}
