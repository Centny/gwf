package timer

import (
	"fmt"
	"sync"
	"time"

	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/util"
)

type TimerF func(uint64) error

func (t TimerF) OnTime(i uint64) error {
	return t(i)
}

func (t TimerF) Name() string {
	return util.FuncName(t)
}

type Timer interface {
	OnTime(i uint64) error
	Name() string
}

var ShowLog = false
var timer_m = map[string]Timer{}  //id map to timer
var timer_d = map[string]int64{}  //id map to delay
var timer_l = map[string]int64{}  //id map to last
var timer_i = map[string]uint64{} //id map to index
var timer_w = map[string]bool{}   //id map to index
var running = map[string]bool{}
var timer_lck = sync.RWMutex{}
var trunning = false

func RegisterV(delay int64, id string, t Timer, wait bool) {
	RegisterV_(delay, id, t, wait, false)
}
func RegisterV_(delay int64, id string, t Timer, wait, run bool) {
	timer_lck.Lock()
	timer_m[id] = t
	timer_d[id] = delay / 100 * 100
	if run {
		timer_l[id] = 0
	} else {
		timer_l[id] = util.Now()
	}
	timer_i[id] = 0
	timer_w[id] = wait
	if !trunning {
		trunning = true
		go loop_timer()
	}
	timer_lck.Unlock()
	log.D("Register timer(%v/%v) success", id, t.Name())
}

func Register(delay int64, t Timer) {
	RegisterV(delay, fmt.Sprintf("%p", t), t, false)
}

func Register2(delay int64, f func(uint64) error) {
	RegisterV(delay, fmt.Sprintf("%p", f), TimerF(f), false)
}

func Register3(delay int64, t Timer, wait bool) {
	RegisterV(delay, fmt.Sprintf("%p", t), t, wait)
}

func Register4(delay int64, f func(uint64) error, wait bool) {
	RegisterV(delay, fmt.Sprintf("%p", f), TimerF(f), wait)
}

func Register5(delay int64, f func(uint64) error, wait, run bool) {
	RegisterV_(delay, fmt.Sprintf("%p", f), TimerF(f), wait, run)
}

func RemoveV(id string) {
	timer_lck.Lock()
	var t = timer_m[id]
	delete(timer_m, id)
	delete(timer_d, id)
	delete(timer_l, id)
	delete(timer_i, id)
	timer_lck.Unlock()
	if t == nil {
		log.W("Register timer(%v) fail with not found", id)
	} else {
		log.D("Register timer(%v/%v) success", id, t.Name())
	}
}

func Remove(t Timer) {
	RemoveV(fmt.Sprintf("%p", t))
}

func Remove2(f func(uint64) error) {
	RemoveV(fmt.Sprintf("%p", f))
}

func Stop() {
	timer_lck.Lock()
	trunning = false
	timer_lck.Unlock()
}

func loop_timer() {
	for trunning {
		timer_lck.RLock()
		now := util.Now()
		for id, t := range timer_m {
			if now-timer_l[id] < timer_d[id] {
				continue
			}
			if timer_w[id] && running[id] {
				continue
			}
			running[id] = true
			timer_l[id] = now
			timer_i[id] += 1
			go run_timer(id, t, timer_i[id])
		}
		timer_lck.RUnlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func run_timer(id string, t Timer, i uint64) {
	defer func() {
		var err = recover()
		if err != nil {
			log.E("timer calling on timer(%v) panic(%v) with stack:\n", t.Name(), err, util.CallStatck())
		}
		timer_lck.Lock()
		running[id] = false
		timer_lck.Unlock()
	}()
	var err = t.OnTime(i)
	if err != nil {
		log.E("timer calling on timer(%v) fail with error(%v)", t.Name(), err)
	} else if ShowLog {
		log.D("timer calling on timer(%v) success", t.Name())
	}
}
