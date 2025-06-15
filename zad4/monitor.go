package main

type responseChannel struct {
	resp chan struct{}
}

type Monitor struct {
	enterChannel     <-chan responseChannel
	leaveChannel     <-chan responseChannel
	terminateChannel <-chan struct{}
}

func (m *Monitor) run() {
	for {
		select {
		case msg := <-m.enterChannel:
			msg.resp <- struct{}{}
		case <-m.terminateChannel:
			return
		}
		select {
		case msg := <-m.leaveChannel:
			msg.resp <- struct{}{}
		case <-m.terminateChannel:
			return
		}
	}
}

type Condition struct {
	myCount        int
	signalChannel  chan responseChannel
	preWaitChannel chan responseChannel
	waitChannel    chan responseChannel
}

func createCondition() *Condition {
	return &Condition{
		myCount:        0,
		signalChannel:  make(chan responseChannel),
		preWaitChannel: make(chan responseChannel),
		waitChannel:    make(chan responseChannel),
	}
}

func (c *Condition) run(monitor *Monitor) {

}
