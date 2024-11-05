package databases

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
)

// Config содержит настройки для подключения к Elasticsearch
type Config struct {
	Addresses []string
	Username  string
	Password  string
}

// Client оборачивает клиент Elasticsearch
type Client struct {
	ES *elasticsearch.Client
}

// NewClient создает новый клиент Elasticsearch
func NewClient(config Config) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating the Elasticsearch client: %v", err)
	}
	return &Client{ES: es}, nil
}
