package main

type responseChannel struct {
	resp chan struct{}
}

type boolResponseChannel struct {
	resp chan bool
}

type Monitor struct {
	enterChannel     chan responseChannel
	leaveChannel     chan responseChannel
	terminateChannel chan struct{}
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
	myCount          int
	signalChannel    chan responseChannel
	preWaitChannel   chan responseChannel
	waitChannel      chan responseChannel
	isWaitingChannel chan boolResponseChannel
	terminateChannel chan struct{}
}

func createCondition() *Condition {
	return &Condition{
		myCount:          0,
		signalChannel:    make(chan responseChannel),
		preWaitChannel:   make(chan responseChannel),
		waitChannel:      make(chan responseChannel),
		isWaitingChannel: make(chan boolResponseChannel),
		terminateChannel: make(chan struct{}),
	}
}

func (c *Condition) run(monitor *Monitor) {
	for {
		var monitorResponseChannel = make(chan struct{})
		if c.myCount == 0 {
			select {
			case msg := <-c.signalChannel:

				monitor.leaveChannel <- responseChannel{resp: monitorResponseChannel}
				<-monitorResponseChannel
				msg.resp <- struct{}{}

			case msg := <-c.preWaitChannel:
				c.myCount++
				msg.resp <- struct{}{}

			case msg := <-c.waitChannel:
				msg.resp <- struct{}{}

			waitLoop:
				for {
					select {
					case msg := <-c.signalChannel:
						c.myCount--
						msg.resp <- struct{}{}
						break waitLoop

					case msg := <-c.preWaitChannel:
						c.myCount++
						msg.resp <- struct{}{}
					case msg := <-c.isWaitingChannel:
						msg.resp <- true
					case <-c.terminateChannel:
						return
					}
				}
			case msg := <-c.isWaitingChannel:
				msg.resp <- c.myCount > 0
			case <-c.terminateChannel:
				return
			}
		} else {
			select {
			case msg := <-c.preWaitChannel:
				c.myCount++
				msg.resp <- struct{}{}

			case msg := <-c.waitChannel:
				msg.resp <- struct{}{}

			waitLoop2:
				for {
					select {
					case msg := <-c.signalChannel:
						c.myCount--
						msg.resp <- struct{}{}
						break waitLoop2

					case msg := <-c.preWaitChannel:
						c.myCount++
						msg.resp <- struct{}{}
					case msg := <-c.isWaitingChannel:
						msg.resp <- true
					case <-c.terminateChannel:
						return
					}
				}
			case msg := <-c.isWaitingChannel:
				msg.resp <- c.myCount > 0
			case <-c.terminateChannel:
				return
			}
		}
	}
}
