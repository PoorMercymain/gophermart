package util

import (
	"encoding/json"
	"errors"
	"strconv"
)

var ErrDuplicate = errors.New("duplicate key in JSON")

func dupErr(path []string) error {
	return ErrDuplicate
}

func CheckDuplicatesInJSON(d *json.Decoder, path []string) error {
	// Get next token from JSON
	t, err := d.Token()
	if err != nil {
		return err
	}

	// Is it a delimiter?
	delim, ok := t.(json.Delim)

	// No, nothing more to check.
	if !ok {
		// scaler type, nothing to do
		return nil
	}

	switch delim {
	case '{':
		keys := make(map[string]bool)
		for d.More() {

			// Get field key.

			t, err := d.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			// Check for duplicates.

			if keys[key] {
				// Duplicate found. Call the application's dup function. The
				// function can record the duplicate or return an error to stop
				// the walk through the document.
				if err := dupErr(append(path, key)); err != nil {
					return err
				}
			}
			keys[key] = true

			// Check value.

			if err := CheckDuplicatesInJSON(d, append(path, key)); err != nil {
				return err
			}
		}
		// consume trailing }
		if _, err := d.Token(); err != nil {
			return err
		}

	case '[':
		i := 0
		for d.More() {
			if err := CheckDuplicatesInJSON(d, append(path, strconv.Itoa(i))); err != nil {
				return err
			}
			i++
		}
		// consume trailing ]
		if _, err := d.Token(); err != nil {
			return err
		}

	}
	return nil
}
