package config

import (
    "encoding/json"
    "os"
    "errors"
)

// Main configuration structure
type Configuration struct {
    PublicDir string `json:"public_dir"`
    TemplateDir string `json:"template_dir"`
    GlobalVars struct {
        SiteTitle string `json:"site_title"`
    } `json:"global_vars"`
    ConnectionString string `json:"connection_string"`
}

func (c *Configuration) Validate() error {
    if c.PublicDir == "" {
        return errors.New("public_dir cannot be empty")
    }

    if c.TemplateDir == "" {
        return errors.New("template_dir cannot be empty")
    }

    if c.GlobalVars.SiteTitle == "" {
        return errors.New("global_vars.site_title cannot be empty")
    }

    if c.ConnectionString == "" {
        return errors.New("connetion_string cannot be empty")
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

func SaveConfiguration() {
    loadConfiguration()

    file, err := os.Open("./goblog.config.json")
    if err != nil {
        panic(err)
    }

    encoder := json.NewEncoder(file)
    err = encoder.Encode(config)
    if err != nil {
        panic(err)
    }
}

func GetConfiguration() *Configuration {
    loadConfiguration()

    return config
}
