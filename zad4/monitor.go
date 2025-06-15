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

func createMonitor() *Monitor {
	return &Monitor{
		enterChannel:     make(chan responseChannel),
		leaveChannel:     make(chan responseChannel),
		terminateChannel: make(chan struct{}),
	}
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
	monitor          *Monitor
}

func createCondition(monitor *Monitor) *Condition {
	return &Condition{
		myCount:          0,
		signalChannel:    make(chan responseChannel),
		preWaitChannel:   make(chan responseChannel),
		waitChannel:      make(chan responseChannel),
		isWaitingChannel: make(chan boolResponseChannel),
		terminateChannel: make(chan struct{}),
		monitor:          monitor,
	}
}

func (c *Condition) run() {
	for {
		var monitorResponseChannel = make(chan struct{})
		if c.myCount == 0 {
			select {
			case msg := <-c.signalChannel:

				c.monitor.leaveChannel <- responseChannel{resp: monitorResponseChannel}
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

func nonEmpty(c *Condition) bool {
	var respChannel = boolResponseChannel{resp: make(chan bool)}
	c.isWaitingChannel <- respChannel
	return <-respChannel.resp
}

func (c *Condition) signal() {
	var respChannel = responseChannel{resp: make(chan struct{})}
	c.signalChannel <- respChannel
	<-respChannel.resp
}
func (c *Condition) preWait() {
	var respChannel = responseChannel{resp: make(chan struct{})}
	c.preWaitChannel <- respChannel
	<-respChannel.resp
}
func (c *Condition) wait() {
	var respChannel = responseChannel{resp: make(chan struct{})}
	c.waitChannel <- respChannel
	<-respChannel.resp
}

func (c *Condition) publicWait() {
	c.preWait()
	var respChannel = make(chan struct{})
	c.monitor.leaveChannel <- responseChannel{resp: respChannel}
	<-respChannel
	c.wait()
}
