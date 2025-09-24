package logo

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	tomb "gopkg.in/tomb.v2"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"

	log2 "log"
)

type LogoCoroutineState int

const (
	LCRStateCreated LogoCoroutineState = iota
	LCRStateRunning
	LCRStatePaused
	LCRStateDying
	LCRStateDead
	LCRStateError
)

func (s LogoCoroutineState) String() string {
	switch s {
	case LCRStateCreated:
		return "created"
	case LCRStateRunning:
		return "running"
	case LCRStatePaused:
		return "paused"
	case LCRStateDying:
		return "stopping"
	case LCRStateDead:
		return "stopped"
	case LCRStateError:
		return "error"
	}
	return "unknown"
}

type LogoCoroutine struct {
	driver      *LogoDriver
	parent      *LogoDriver
	t           *tomb.Tomb
	err         error
	CoroutineID string
	Command     *types.TokenList
	State       LogoCoroutineState
}

var crid int64 = 1

func getCoroutineId() string {
	atomic.AddInt64(&crid, 1)
	return fmt.Sprintf("%d", crid)
}

// NewLogoCoroutine creates a new instance of a LogoCoroutine, running concurrently
// to the main intepreter.
func NewLogoCoroutine(parent *LogoDriver) *LogoCoroutine {
	l := &LogoCoroutine{
		parent:      parent,
		t:           &tomb.Tomb{},
		driver:      NewLogoDriverInherit(parent),
		CoroutineID: getCoroutineId(),
		State:       LCRStateCreated,
	}
	l.driver.ent = parent.ent
	return l
}

func (d *LogoDriver) GetCoroutine(id string) (*LogoCoroutine, bool) {
	d.crt.Lock()
	defer d.crt.Unlock()
	c, ok := d.Coroutines[id]
	return c, ok
}

func (d *LogoDriver) WaitAllCoroutines() error {
	for len(d.Coroutines) > 0 && !d.HasCBreak() && !d.HasReboot() {
		time.Sleep(5 * time.Millisecond)
	}
	return nil
}

// KillAllCoroutines kills all coroutines, even those spawned by the coroutine
func (d *LogoDriver) KillAllCoroutines() error {

	crlist := []string{}
	for k, _ := range d.Coroutines {
		crlist = append(crlist, k)
	}

	for _, id := range crlist {
		cr := d.Coroutines[id]
		if cr != nil {
			cr.Stop()
		}
	}

	return nil
}

func (d *LogoDriver) KillCoroutine(name string) error {
	if cr, ok := d.Coroutines[name]; ok {
		cr.Stop()
		return nil
	}
	return errors.New("no logoroutine with id " + name)
}

func (d *LogoDriver) RegisterCoroutine(c *LogoCoroutine) {
	d.crt.Lock()
	defer d.crt.Unlock()
	d.Coroutines[c.CoroutineID] = c
}

func (d *LogoDriver) DeregisterCoroutine(c *LogoCoroutine) {
	d.crt.Lock()
	defer d.crt.Unlock()
	delete(d.Coroutines, c.CoroutineID)
}

func (d *LogoDriver) SpawnCoroutine(command *types.TokenList) (string, error) {

	if command == nil {
		return "", errors.New("loGO routine needs a list")
	}

	c := NewLogoCoroutine(d)
	d.RegisterCoroutine(c)
	err := c.Start(command)
	return c.CoroutineID, err
}

func (c *LogoCoroutine) IsRunning() bool {
	switch c.State {
	case LCRStateRunning, LCRStatePaused:
		return true
	default:
		return false
	}
	return false
}

func (c *LogoCoroutine) Purge() error {
	if c.State == LCRStateDead || c.State == LCRStateError || c.State == LCRStateCreated {
		c.parent.DeregisterCoroutine(c)
		return nil
	}
	return errors.New("coroutine not stopped")
}

func (c *LogoCoroutine) Stop() {
	if c.State != LCRStateDead && c.State != LCRStateError && c.State != LCRStateCreated {
		log2.Printf("Killing coroutine with id %s", c.CoroutineID)
		c.t.Kill(errors.New("stopped"))
	}
	c.State = LCRStateDead
	c.parent.DeregisterCoroutine(c)
}

func (c *LogoCoroutine) Start(command *types.TokenList) error {
	if c.State != LCRStateCreated {
		return errors.New("cannot start logoroutine")
	}
	c.Command = command.Copy()
	c.t.Go(c.Run)
	return nil
}

func (c *LogoCoroutine) Run() error {
	c.State = LCRStateRunning

	frame := c.driver.Stack.Size()

	c.driver.ReresolveSymbols(c.driver.d.Lexer)
	err := c.driver.CreateCoroutineScope(c.Command, c.CoroutineID)
	if err != nil {
		c.State = LCRStateError
		c.err = err
		apple2helpers.OSDShow(c.driver.ent, fmt.Sprintf("Error: %v (in %s)", err, c.CoroutineID))
		c.Purge()
		return err
	}

	_, err = c.driver.ExecTillStackLevelReachedWithTomb(frame, c.t)
	if err != nil {
		c.State = LCRStateError
		c.err = err
		log2.Printf("Error executing coroutine: %v", err)
		apple2helpers.OSDShow(c.driver.ent, fmt.Sprintf("Error: %v (in %s)", err, c.CoroutineID))
		c.Purge()
		return err
	}

	c.State = LCRStateDead
	c.err = nil

	c.Purge()

	return nil
}

func randomString(l int) string {
	out := ""
	set := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for len(out) < l {
		out += string(rune(set[rand.Intn(len(set))]))
	}
	return out
}
