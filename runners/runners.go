package runners

import (
	"sync"
)

// InitializeRunners starts necessary runners as threads
func InitializeRunners() {
	var wg sync.WaitGroup
	wg.Add(3) // terminate when all threads terminate

	go InitializeLookup(&wg)
	go InitializeCollect(&wg)
	go InitializePrinter(&wg)

	wg.Wait()
}
