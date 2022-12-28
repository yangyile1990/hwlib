package timer

import (
	"fmt"
	"sync"
	"time"
)

const SEC int = 0
const MIN int = 1
const HOUR int = 2

type Do func()
type task struct {
	hour    int
	min     int
	sec     int
	repeat  bool
	handler Do
	dur     time.Duration
}

//一个时间轮
type Ticker struct {
	sync.Mutex
	degree     time.Duration
	slots      [][]*task //任务列
	currentIdx int
	maxIdx     int
}

func (t *Ticker) addTask(dur int, input *task) {
	t.Lock()
	defer t.Unlock()
	idx := dur % t.maxIdx
	if t.slots[idx] == nil {
		t.slots[idx] = make([]*task, 0)
	}
	t.slots[idx] = append(t.slots[idx], input)
}

//工作
func (t *Ticker) tick() ([]*task, bool) {
	newRound := false
	t.Lock()
	defer t.Unlock()
	t.currentIdx++
	if t.currentIdx >= t.maxIdx {
		newRound = true
		t.currentIdx = t.currentIdx % t.maxIdx
	}
	out := t.slots[t.currentIdx]
	t.slots[t.currentIdx] = make([]*task, 0)
	return out, newRound
}
func newTicker(degree time.Duration, slots int, now time.Time) *Ticker {
	t := &Ticker{degree: degree,
		slots:      make([][]*task, slots),
		currentIdx: 0,
		maxIdx:     slots,
	}
	return t
}

type Timer struct {
	sec      *Ticker
	min      *Ticker
	hour     *Ticker
	stop     chan bool
	lastTick time.Time
}

func NewTimer() *Timer {
	now := time.Now()
	t := &Timer{
		sec:  newTicker(time.Second, 60, now),
		min:  newTicker(time.Minute, 60, now),
		hour: newTicker(time.Hour, 24, now),
		stop: make(chan bool, 1),
	}
	return t
}

func (t *Timer) AddTask(dur time.Duration, repeat bool, handler Do) error { //是否重复
	if dur < time.Second {
		return fmt.Errorf("min 1s")
	}
	if dur >= time.Hour*24 {
		return fmt.Errorf("max 24h")
	}
	tmp := t.sec.currentIdx + int(dur/time.Second)
	sec := tmp % 60
	min := ((tmp/60)%60 + t.min.currentIdx) % 60
	hour := (tmp/3600 + t.hour.currentIdx) % 24
	elem := &task{hour: hour, min: min, sec: sec, repeat: repeat, dur: dur, handler: handler}
	t.addTick(elem)
	return nil
}
func (t *Timer) addTask(elem *task) { //是否重复
	dur := elem.dur
	tmp := t.sec.currentIdx + int(dur/time.Second)
	sec := tmp % 60
	min := ((tmp/60)%60 + t.min.currentIdx) % 60
	hour := (tmp/3600 + t.hour.currentIdx) % 24
	elem.hour = hour
	elem.min = min
	elem.sec = sec
	t.addTick(elem)
}
func (t *Timer) addTick(elem *task) {
	if elem.sec != 0 {
		t.sec.addTask(elem.sec, elem)
	} else if elem.min != 0 {
		t.min.addTask(elem.min, elem)
	} else if elem.hour != 0 {
		t.hour.addTask(elem.hour, elem)
	}
}
func (t *Timer) Run() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 50):
				now := time.Now()
				if now.Sub(t.lastTick) < time.Second {
					continue
				}
				t.lastTick = now
				secTask, up := t.sec.tick()
				go t.doTask(SEC, secTask)
				if !up {
					continue
				}
				minTask, up := t.min.tick()
				go t.doTask(MIN, minTask)
				if !up {
					continue
				}
				hourTask, _ := t.hour.tick()
				go t.doTask(HOUR, hourTask)
			case <-t.stop:
				wg.Done()
			}
		}
	}()
	return wg
}
func (t *Timer) Stop() {
	t.stop <- true
}
func (t *Timer) doTask(lv int, tasks []*task) {
	for _, elem := range tasks {
		if elem.min != 0 && (lv == SEC) {
			t.min.addTask(elem.min, elem)
			continue
		}
		if elem.hour != 0 && (lv == SEC || lv == MIN) {
			t.hour.addTask(elem.hour, elem)
			continue
		}
		if elem.repeat {
			t.addTask(elem)
		}
		elem.handler()
	}
}
