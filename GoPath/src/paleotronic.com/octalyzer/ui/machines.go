package ui

import (
	"sort"
	"strings"

	"paleotronic.com/log"

	yaml "gopkg.in/yaml.v2"
	"paleotronic.com/core/hardware"
	"paleotronic.com/files"
	"paleotronic.com/octalyzer/assets"
)

type MachineInfo struct {
	Name     string
	Filename string
	Order    int
}

type byOrder []MachineInfo

func (s byOrder) Len() int {
	return len(s)
}

func (s byOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byOrder) Less(i, j int) bool {
	return s[i].Order < s[j].Order
}

func GetMachineList() []MachineInfo {

	m := byOrder{}

	filelist, err := assets.AssetDir("profile")
	log.Printf("asset dir err: %v", err)
	if err != nil {
		return m
	}

	log.Printf("filelist = %+v", filelist)

	for _, f := range filelist {
		if strings.HasSuffix(f, ".yaml") {
			data, err := assets.Asset("profile/" + f)
			if err != nil {
				continue
			}
			var s hardware.MachineSpec
			err = yaml.Unmarshal(data, &s)
			if err != nil {
				continue
			}
			machine := MachineInfo{
				Name:     s.Name,
				Order:    s.SortOrder,
				Filename: files.GetFilename(f),
			}
			m = append(m, machine)
		}
	}

	log.Printf("m = %+v", m)

	sort.Sort(m)

	return m

}
