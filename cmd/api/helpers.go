package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type envelope map[string]any

func (app *app) readIdParam(r *http.Request) (int64, error) {
	strId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(strId, 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	res, err := json.Marshal(data)

	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
	return nil
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1048756 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains bad JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains bad JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body cant be empty")

		case strings.HasPrefix(err.Error(), "json: Unkown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: Unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must be less than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = decoder.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
