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

	startSymbol := 'A'

	processes := make([]Process, Nr_Of_Processes)
	for i := 0; i < Nr_Of_Processes; i++ {
		processes[i].initProcess(i, rand.Int(), startSymbol+int32(i))
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
	fmt.Println("LOCAL_SECTION;ENTRY_PROTOCOL1;ENTRY_PROTOCOL2;ENTRY_PROTOCOL3;ENTRY_PROTOCOL4;CRITICAL_SECTION;EXIT_PROTOCOL;")
}
