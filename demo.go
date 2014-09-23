package main

import (
	"fmt"
	"time"
)

const count = 100

func main() {
	fmt.Printf("\n** mutable state **\n")

	blockingPull()

	//blockingPush()

	//stackChanPull()

	//stackChanPush()

	//defaultChan()

	//closedChan()

	//simple()

	//shared_concurrent()

	//coordinated_concurrent()

	fmt.Printf("\n\n")
	for {
		<-time.After(1 * time.Second)
		fmt.Printf("*")
	}

}

func blockingPull() {
	blockingChan := make(chan struct{})

	<-blockingChan
}

func blockingPush() {
	blockingChan := make(chan struct{})

	blockingChan <- struct{}{}
}

func stackChanPull() {
	stackChan := make(chan struct{}, 1)

	<-stackChan
}

func stackChanPush() {
	stackChan := make(chan struct{}, 1)

	anon := func() {
		for {
			fmt.Printf("\n\nDo stuff...\n\n\n")

			<-time.After(2 * time.Second)
		}
	}

	go anon()

	stackChan <- struct{}{}
	stackChan <- struct{}{}
}

func defaultChan() {
	blockingChanLeft := make(chan struct{})
	blockingChanRight := make(chan struct{})

	for {
		select {
		case <-blockingChanLeft:
			{
				fmt.Printf("Should have blocked!!\n")
			}
		case <-blockingChanRight:
			{
				fmt.Printf("Should have blocked 2!!\n")
			}
		default:
			{
				fmt.Printf("Was blocked :)\n")
			}
		}
	}
}

func closedChan() {
	blockingChan := make(chan struct{})
	cancelChan := make(chan struct{})

	go func() {
		<-time.After(4 * time.Second)
		close(cancelChan)
	}()

	for {
		timer := time.After(200 * time.Millisecond)
		select {
		case <-blockingChan:
			{
				fmt.Printf("Should have blocked\n")
				return
			}
		case <-cancelChan:
			{
				fmt.Printf("Canceled!!\n")
				return
			}
		case <-timer:
			{
				fmt.Printf("Still blocked...\n")
			}
		}
	}
}

func simple() {
	queue := make([]int, 0, 10)

	// Push
	queue = append(queue, 1)
	queue = append(queue, 2)
	queue = append(queue, 3)

	fmt.Printf("\nPrint items...\n\n")
	for index, item := range queue {
		fmt.Printf("Items[%d]: [ %d ]\n", index, item)
	}

	fmt.Printf("\nPop all the items...\n\n")
	// Pop
	fmt.Printf("Value: [ %d ]\n", queue[0])
	queue = queue[1:]
	fmt.Printf("Value: [ %d ]\n", queue[0])
	queue = queue[1:]
	fmt.Printf("Value: [ %d ]\n", queue[0])
	queue = queue[1:]

	fmt.Printf("\nPrint items...\n\n")
	for index, item := range queue {
		fmt.Printf("Items[%d]: [ %d ]\n", index, item)
	}
	fmt.Printf("\nDone!\n")
}

func shared_concurrent() {
	// Shared state
	queue := make([]int, 0, 100)

	for i := 0; i < count; i++ {
		go func(index int) {
			go func() {
				// Push goroutine
				fmt.Printf("Push value [ %d ]\n\n", index)
				queue = append(queue, index)
			}()
			go func() {
				// Print goroutine
				fmt.Printf("\nPrint items async...\n")
				for index, item := range queue {
					fmt.Printf("Items[%d]: [ %d ]\n", index, item)
				}
				fmt.Printf("\n")
			}()
			go func() {
				// Pop
				if len(queue) > 0 {
					value := queue[0]
					fmt.Printf("Pop value [ %d ]\n\n", value)
					queue = queue[1:]
				}
			}()
		}(i)
	}
	fmt.Printf("\nPrint items...\n\n")
	for index, item := range queue {
		fmt.Printf("Items[%d]: [ %d ]\n", index, item)
	}
}

func coordinated_concurrent() {
	// Coordinating channels
	pushChan := make(chan int)
	popChan := make(chan chan int)
	printChan := make(chan struct{})
	pauseChan := make(chan chan struct{})

	// Managed state
	go func() {
		queue := make([]int, 0, 100)

		for {
			select {
			case value := <-pushChan:
				{
					fmt.Printf("Push value [ %d ]\n\n", value)
					queue = append(queue, value)
				}
			case popper := <-popChan:
				//case popChan <- 5:
				{
					if len(queue) > 0 {
						value := queue[0]
						fmt.Printf("Pop value [ %d ]\n\n", value)
						popper <- value
						queue = queue[1:]
					}
					// Broken forever!!
					close(popper)
				}
			case <-printChan:
				{
					fmt.Printf("\nPrint items chan...\n")
					for index, item := range queue {
						fmt.Printf("Items[%d]: [ %d ]\n", index, item)
					}
					fmt.Printf("\n")
				}
			case pauser := <-pauseChan:
				{
					fmt.Printf("Paused...")
					<-pauser
					fmt.Printf("\n\n")
				}
			}
		}
	}()

	for i := 0; i < count; i++ {
		go func(index int) {
			// Push goroutine
			go func() {
				pushChan <- index
			}()
			// Print goroutine
			go func() {
				printChan <- struct{}{}
			}()
			// Pop goroutine
			go func() {
				popper := make(chan int)
				popChan <- popper
				<-popper // Never got called
			}()
			// Pause goroutine
			go func() {
				pauser := make(chan struct{})
				pauseChan <- pauser
				// Do some stuff while the queue is suspended
				<-time.After(1 * time.Second)

				pauser <- struct{}{}
			}()
		}(i)
	}
}
