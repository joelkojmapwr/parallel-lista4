package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func main() {
	printerChannel := make(chan string, 1)

	var processWait sync.WaitGroup

	done := make(chan struct{})

	go printer(printerChannel, done)

	rwMonitor := createRWMonitor()
	go rwMonitor.run()
	defer rwMonitor.terminate()

	processes := make([]Process, Nr_Of_Processes)
	for i := 0; i < Nr_Of_Readers; i++ {
		processes[i].initProcess(i, rand.Int(), 'R', rwMonitor)
	}

	for i := 0; i < Nr_Of_Writers; i++ {
		processes[i+Nr_Of_Readers].initProcess(i+Nr_Of_Readers, rand.Int(), 'W', rwMonitor)
	}

	for i := 0; i < Nr_Of_Processes; i++ {
		processWait.Add(1)
		go processes[i].run(printerChannel, &processWait)
	}

	// printerChannel <- LocalSection.String()
	processWait.Wait()

	close(printerChannel)
	<-done
	fmt.Print(-1, " ", Nr_Of_Processes, " ", Board_Width, " ", Board_Height, " ")
	// Assuming you have a variable named processStates of type []ProcessState or map[int]ProcessState
	// for i, v := range processStates {
	// 	fmt.Printf("%d: %s\n", i, v)
	// }
	fmt.Println("LOCAL_SECTION;START;READING_ROOM;STOP;")
}
