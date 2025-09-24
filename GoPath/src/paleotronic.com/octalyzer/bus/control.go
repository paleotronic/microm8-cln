package bus

import (
	"sync"
	"time"

	"paleotronic.com/core/settings"
	"paleotronic.com/fmt"
)

var CPUControl bool
var ClockTime time.Duration = time.Second / 60
var SyncNeeded, SyncComplete bool
var lock sync.Mutex
var SyncCallback func()
var lastSync time.Time

func init() {
	ticker = time.NewTicker(ClockTime)
	go func() {
		for {
			select {
			case <-ticker.C:
				if running && time.Since(lastSync) >= ClockTime {
					Sync()
				}
			}
		}
	}()
}

var ticker *time.Ticker
var running bool
var quit chan bool

func StartDefault() {
	StartClock(ClockTime)
}

func StartClock(d time.Duration) {

	if settings.IsRemInt {
		return
	}

	ClockTime = d

	running = true

	fmt.Println("Start the clock...")

}

func StopClock() {

	if settings.IsRemInt {
		return
	}

	running = false
	fmt.Println("Stop the clock...")
}

func Sync() {

	if settings.IsRemInt {
		return
	}

	lastSync = time.Now()

	lock.Lock()
	defer lock.Unlock()

	if SyncCallback != nil {
		SyncCallback()
	}
}

func SetCallback(f func()) {
	if settings.IsRemInt {
		return
	}

	SyncCallback = f
}

func SyncDo(f func()) {
	if settings.IsRemInt {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	///fmt.Println("syncDo")

	if f != nil {
		f()
	}
}

func IsClock() bool {
	return running
}
