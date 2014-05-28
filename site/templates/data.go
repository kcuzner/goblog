package templates

import (
    "bitbucket.org/kcuzner/goblog/site/config"
)

// Global template data
type GlobalVars struct {
    SiteTitle string
}

func makeGlobalVars() GlobalVars {
    c := config.GetConfiguration()

    return GlobalVars{c.GlobalVars.SiteTitle}
}

var globalVars = makeGlobalVars()

// Gets the global variables for templates
func GetGlobalVars() GlobalVars {
    return globalVars
}

