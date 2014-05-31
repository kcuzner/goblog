package templates

import (
    "bitbucket.org/kcuzner/goblog/site/config"
)

type FlashKey string

const (
    ErrorFlashKey FlashKey = "error"
    WarningFlashKey = "warning"
    InfoFlashKey = "info"
    SuccessFlashKey = "success"
)

// Source for flashes compatible with gorrila sessions
type FlashSource interface {
    Flashes(...string) []interface{}
}

// Store for flashes compatible with gorrilla sessions
type FlashStore interface {
    AddFlash(value interface{}, vars ...string)
}

// Global template data
type GlobalVars struct {
    SiteTitle string
    Errors, Warnings, Infos, Successes []string
}

var configuration = config.GetConfiguration()

func Flash(s FlashStore, key FlashKey, value string) {
    s.AddFlash(value, string(key))
}

func getFlashes(s FlashSource, key FlashKey) []string {
    inFlashes := s.Flashes(string(key))
    outFlashes := make([]string, 0)
    for i := range inFlashes {
        outFlashes = append(outFlashes, inFlashes[i].(string))
    }

    println(key, outFlashes)

    return outFlashes
}

// Gets the global variables for templates
func GetGlobalVars(s FlashSource) GlobalVars {
    return GlobalVars{configuration.GlobalVars.SiteTitle,
        getFlashes(s, ErrorFlashKey),
        getFlashes(s, WarningFlashKey),
        getFlashes(s, InfoFlashKey),
        getFlashes(s, SuccessFlashKey)}
}

