package config

import (
    "encoding/json"
    "os"
    "errors"
)

// Main configuration structure
type configuration struct {
    PublicDir string `json:"public_dir"`
    TemplateDir string `json:"template_dir"`
    GlobalVars struct {
        SiteTitle string `json:"site_title"`
    } `json:"global_vars"`
    ConnectionString string `json:"connection_string"`
}

func (c *configuration) validate() error {
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

func (c *configuration) Save() {
    file, err := os.Open("./goblog.config.json")
    if err != nil {
        panic(err)
    }

    encoder := json.NewEncoder(file)
    err = encoder.Encode(c)
    if err != nil {
        panic(err)
    }
}

func loadConfiguration() *configuration {
    file, err := os.Open("./goblog.config.json")
    if err != nil {
        panic(err)
    }

    config := &configuration{}
    decoder := json.NewDecoder(file)
    err = decoder.Decode(config)
    if err != nil {
        panic(err)
    }

    err = config.validate()
    if err != nil {
        panic(err)
    }

    return config
}

var Config *configuration = loadConfiguration()
