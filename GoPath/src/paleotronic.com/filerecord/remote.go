package filerecord

type RemoteInstance struct {
	Host string
	Port int
	Owner string
	PID  int
	CheckSum [16]byte
	Freeze []byte
	UUID string
    // ACLs for remote instance
    ViewGroups  []string
    UseGroups   []string
    AdminGroups []string
}
