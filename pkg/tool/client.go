package tool

import (
	"github.com/gaebalai/secfarm/pkg/adapter"
	"github.com/gaebalai/secfarm/pkg/repository"
)

// Client contains shared resources that tools can use
type Client struct {
	Repo    repository.Repository
	Gemini  adapter.Gemini
	Storage adapter.Storage
}
