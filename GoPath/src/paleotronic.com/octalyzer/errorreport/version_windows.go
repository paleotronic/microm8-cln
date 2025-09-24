package errorreport

import "golang.org/x/sys/windows/registry"

func GetOSVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Unknown"
	}
	defer k.Close()
	pn, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return "Unknown"
	}
	return pn
}
