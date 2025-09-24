package files

import (
	"paleotronic.com/api"
	"paleotronic.com/log"
)

func ProjectMapFunc() map[string]ProviderHolder {
	m := make(map[string]ProviderHolder)
	
	// get logged in users projects
	log.Println("Calling FetchProjectList()")
	list, e := s8webclient.CONN.FetchProjectList()
	log.Println(e)
	log.Println(list)
	if e != nil {
		return m
	}
	
	// build list using networkproviders
	for _, p := range list {
		m[p] = ProviderHolder{ BasePath: p, Provider: NewProjectProvider("", true, false, true, p, 0) }
	}
	
	log.Println(m)
	
	return m
}