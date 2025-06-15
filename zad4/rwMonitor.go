package main

type RWMonitor struct {
	OK_to_Read       *Condition
	OK_to_Write      *Condition
	writing          bool
	terminateChannel chan struct{}
	readers          int
	monitor          *Monitor
}

func createRWMonitor() *RWMonitor {
	monitor1 := createMonitor()
	return &RWMonitor{
		OK_to_Read:       createCondition(monitor1),
		OK_to_Write:      createCondition(monitor1),
		writing:          false,
		readers:          0,
		terminateChannel: make(chan struct{}),
		monitor:          monitor1,
	}
}

func (rw *RWMonitor) run() {
	go rw.monitor.run()
	go rw.OK_to_Read.run()
	go rw.OK_to_Write.run()

}

func (rw *RWMonitor) terminate() {
	rw.monitor.terminateChannel <- struct{}{}
	rw.OK_to_Read.terminateChannel <- struct{}{}
	rw.OK_to_Write.terminateChannel <- struct{}{}
}

func (rw *RWMonitor) startRead() {
	respChan := make(chan struct{})
	rw.monitor.enterChannel <- responseChannel{resp: respChan}
	<-respChan
	if rw.writing || nonEmpty(rw.OK_to_Write) {
		rw.OK_to_Read.publicWait()
	}
	rw.readers++
	rw.OK_to_Read.signal()
}

func (rw *RWMonitor) stopRead() {
	respChan := make(chan struct{})
	rw.monitor.enterChannel <- responseChannel{resp: respChan}
	<-respChan
	rw.readers--
	if rw.readers == 0 {
		rw.OK_to_Write.signal()
	} else {
		rw.monitor.leaveChannel <- responseChannel{resp: respChan}
		<-respChan
	}

}

func (rw *RWMonitor) startWrite() {
	respChan := make(chan struct{})
	rw.monitor.enterChannel <- responseChannel{resp: respChan}
	<-respChan
	if rw.writing || rw.readers > 0 {
		rw.OK_to_Write.publicWait()
	}
	rw.writing = true
	rw.monitor.leaveChannel <- responseChannel{resp: respChan}
	<-respChan
}

func (rw *RWMonitor) stopWrite() {
	respChan := make(chan struct{})
	rw.monitor.enterChannel <- responseChannel{resp: respChan}
	<-respChan
	rw.writing = false
	if nonEmpty(rw.OK_to_Read) {
		rw.OK_to_Read.signal()
	} else {
		rw.OK_to_Write.signal()
	}
}
