// +build: !remint

package main

type SetWindowRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type KeyRequest struct {
	Key       int `json:"key"`
	ScanCode  int `json:"scancode"`
	Action    int `json:"action"`
	Modifiers int `json:"modifiers"`
}

type InsertDiskRequest struct {
	Drive    int    `json:"drive"`    // 0/1 = disk ii, 2 = smartport
	Filename string `json:"filename"` // leave filename "" for blank image (0/1 only)
}

type GeneralDiskRequest struct {
	Drive  int    `json:"drive"`  // 0/1 = disk ii, 2 = smartport
	Action string `json:"action"` // leave filename "" for blank image (0/1 only)
}

type SettingsUpdateRequest struct {
	Path    string `json:"path,omitempty"`
	Value   string `json:"value,omitempty"`
	Persist bool   `json:"persist,omitempty"`
}

type MouseRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type FreezeRequest struct {
	Path string `json:"path,omitempty"`
}

type ScreenRequest struct {
	Path string `json:"path,omitempty"`
}

type LaunchRequest struct {
	WorkingDir string   `json:"workingDir,omitempty"`
	Disks      []string `json:"disks,omitempty"`
	Pakfile    string   `json:"pakfile,omitempty"`
	SmartPort  string   `json:"smartport,omitempty"`
	RunFile    string   `json:"runfile,omitempty"`
	RunCommand string   `json:"runcommand,omitempty"`
	Dialect    string   `json:"dialect,omitempty"`
}
