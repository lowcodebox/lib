package models

type Codebase struct {
	Provider string                       `envconfig:"PROVIDER" default:"" toml:"Provider"`
	Connect  map[string]map[string]string `toml:"Connect"`
	Embed    map[string]string            `toml:"Embed"`
	S3       map[string]string            `toml:"S3"`
	Local    map[string]string            `toml:"Local"`
	Azure    map[string]string            `toml:"Azure"`
	Git      map[string]string            `toml:"Git"`
	Api      map[string]string            `toml:"Api"`
}
