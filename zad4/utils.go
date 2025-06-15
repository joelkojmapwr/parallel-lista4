package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const Nr_Of_Readers int = 10
const Nr_Of_Writers int = 5
const Nr_Of_Processes int = Nr_Of_Readers + Nr_Of_Writers

const Min_Steps int = 10
const Max_Steps int = 30

const Min_Delay float32 = 0.01
const Max_Delay float32 = 0.15

const Board_Width int = Nr_Of_Processes
const Board_Height int = 4

var Start_Time = time.Now()

type Position struct {
	x int
	y int
}

type Trace_Type struct {
	timestamp time.Duration
	id        int
	position  Position
	symbol    rune
}

type ProcessState int

const (
	LocalSection ProcessState = iota
	Start
	Reading_Room
	Stop
)

func (p ProcessState) String() string {
	return [...]string{"LocalSection", "Start", "Reading_Room", "Stop"}[p]
}

func printer(printerChannel <-chan string, done chan<- struct{}) {

	for msg := range printerChannel {
		fmt.Println(msg)
	}
	done <- struct{}{}
}

func print_Trace(trace Trace_Type) string {
	return fmt.Sprintf(" %.9f  %d  %d  %d  %c", trace.timestamp.Seconds(), trace.id, trace.position.x, trace.position.y, trace.symbol)
}

func isInArray(array []uint32, value uint32) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

type Process struct {
	id        int
	symbol    rune
	position  Position
	traces    []Trace_Type
	generator *rand.Rand
	isReader  bool
	rw        *RWMonitor
}

func (p *Process) changeState(newState ProcessState) {
	timeStamp := time.Since(Start_Time)
	p.position.y = int(newState) // Update position.y to reflect the state
	p.storeTrace(timeStamp)
}

func (p *Process) storeTrace(timestamp time.Duration) {
	newTrace := Trace_Type{
		timestamp: timestamp,
		id:        p.id,
		position:  p.position,
		symbol:    p.symbol,
	}
	p.traces = append(p.traces, newTrace)
}

func (p *Process) initProcess(id int, seed int, symbol rune, monitor *RWMonitor) {
	p.id = id
	p.symbol = symbol
	p.generator = rand.New(rand.NewSource(int64(seed)))
	if symbol == 'R' {
		p.isReader = true
	} else {
		p.isReader = false
	}
	p.rw = monitor
	// Initialize position randomly
	p.position.x = id                // Example randomization
	p.position.y = int(LocalSection) // Example randomization

	// Store initial position
	p.storeTrace(time.Duration(0))
}

func (p *Process) run(printerChannel chan<- string, processWait *sync.WaitGroup) {
	defer processWait.Done()
	Nr_Of_Steps := Min_Steps + p.generator.Intn(Max_Steps-Min_Steps+1)

	for step := 0; step < Nr_Of_Steps/4+1; step++ {
		delay := Min_Delay + rand.Float32()*(Max_Delay-Min_Delay)
		time.Sleep(time.Duration(delay * float32(time.Second)))

		p.changeState(Start)
		if p.isReader {
			p.rw.startRead()
			p.changeState(Reading_Room)
			delay := Min_Delay + rand.Float32()*(Max_Delay-Min_Delay)
			time.Sleep(time.Duration(delay * float32(time.Second)))
			p.changeState(Stop)
			p.rw.stopRead()
		} else {
			p.rw.startWrite()
			p.changeState(Reading_Room)
			delay := Min_Delay + rand.Float32()*(Max_Delay-Min_Delay)
			time.Sleep(time.Duration(delay * float32(time.Second)))
			p.changeState(Stop)
			p.rw.stopWrite()
		}

		p.changeState(LocalSection)
	}

	for trace := range p.traces {
		printerChannel <- print_Trace(p.traces[trace])
	}
}
