package http

import "net/http"

func New(opts ...func(*http.Client) error) (*http.Client, error) {
	client := &http.Client{}
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}
