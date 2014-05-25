package config

import (
    "encoding/json"
    "os"
    "errors"
)

// Main configuration structure
type Configuration struct {
    PublicDir string `json:"public_dir"`
}

func (c *Configuration) Validate() error {
    if c.PublicDir == "" {
        return errors.New("public_dir cannot be empty")
    }

    return nil
}

var config *Configuration

func loadConfiguration() {
    if config != nil {
        return
    }

    file, err := os.Open("./goblog.config.json")
    if err != nil {
        panic(err)
    }

    config = &Configuration{}
    decoder := json.NewDecoder(file)
    err = decoder.Decode(config)
    if err != nil {
        panic(err)
    }

    err = config.Validate()
    if err != nil {
        panic(err)
    }
}

func GetConfiguration() *Configuration {
    loadConfiguration()

    return config
}
