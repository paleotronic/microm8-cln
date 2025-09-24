// entity.go
package types

type EntityState int

const (
	INITIALIZE EntityState = iota
	RUNNING
	DIRECTRUNNING
	STOPPED
	EMPTY
	FUNCTIONRUNNING
	PAUSED
	INPUT
	BREAK
	EXEC6502
	DIRECTEXEC6502
	REMOTE
	PLAYERSTART
	PLAYING
	EXECZ80
	DIRECTEXECZ80
)

func (es EntityState) String() string {
	switch es {
	case RUNNING:
		return "running"
	case DIRECTRUNNING:
		return "running interactive"
	case STOPPED:
		return "stopped"
	case EMPTY:
		return "empty"
	case FUNCTIONRUNNING:
		return "running function"
	case PAUSED:
		return "paused"
	case INPUT:
		return "input"
	case BREAK:
		return "break"
	case EXEC6502:
		return "exec 6502"
	case DIRECTEXEC6502:
		return "exec 6502 interactive"
	case EXECZ80:
		return "exec Z80"
	case DIRECTEXECZ80:
		return "exec Z80 interactive"
	case REMOTE:
		return "remote"
	case PLAYERSTART:
		return "player starting"
	case PLAYING:
		return "playing"
	}

	return "unknown"
}

type EntitySubState int

const (
	ESS_INIT EntitySubState = iota
	ESS_SLEEP
	ESS_EXEC
	ESS_DONE
	ESS_BREAK
)
