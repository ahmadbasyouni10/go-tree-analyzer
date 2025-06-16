package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/tour/tree"
)

type DiffEvent struct {
	Type     string
	Value1   int
	Value2   int
	Position int
}

type Metrics struct {
	NodesWalked      int
	DiffsFound       int
	SameOpsTotal     int
	DiffTreeOpsTotal int
	mu               sync.Mutex
}

// metrics with sync.Mutex to avoid racing and use echo to have a api that shows the metrics on a web page
var appMetrics = Metrics{}

// Walk traverses in inorder then sends each value into the channel
func Walk(ctx context.Context, t *tree.Tree, ch chan int) {
	if t == nil {
		return
	}

	appMetrics.mu.Lock()
	appMetrics.NodesWalked++
	appMetrics.mu.Unlock()

	select {
	// if context is done, return early (only will happen if context is cancelled) so channel closed and ready to be read
	case <-ctx.Done():
		fmt.Printf("Walk cancelled for node: %v (before left child): %v\n", t.Value, ctx.Err())
		return
	// this is what makes the check nonblocking
	default:
		// continue normal execution
		fmt.Printf("Walking node: %v\n", t.Value)
	}

	time.Sleep(5 * time.Millisecond) // delaying so i can see the timeout context work
	Walk(ctx, t.Left, ch)

	select {
	case <-ctx.Done():
		fmt.Printf("Walk cancelled for node: %v (before sending value): %v\n", t.Value, ctx.Err())
		return
	case ch <- t.Value:
		fmt.Printf("Value sent success: %v\n", t.Value)
	}

	time.Sleep(5 * time.Millisecond)

	Walk(ctx, t.Right, ch)
}

// Same checks if two trees have the same values in order traversal
func Same(t1, t2 *tree.Tree) bool {

	appMetrics.mu.Lock()
	appMetrics.SameOpsTotal++
	appMetrics.mu.Unlock()

	ch1 := make(chan int)
	ch2 := make(chan int)
	// use waitgroup.add(1) outside of go routine if using wait groups
	// wait after both routines can be called before incrementing and think everything is done
	// so go routine might not finish before the main function exits if inside of rotuine
	go func() {
		Walk(context.Background(), t1, ch1)
		// let the receiver know that the channel is closed
		close(ch1)
	}()

	go func() {
		// defer wg.done in the go routine not in the recurisve call
		// but here i just use channels no wait group bc
		// concurrently need to be checking values as they come in to the
		// channel, if i waited i would be not using concurrency
		Walk(context.Background(), t2, ch2)
		close(ch2)
	}()

	for {
		// receiver for both channels so no deadlock occurs
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2

		if ok1 != ok2 || v1 != v2 {
			return false
		}
		if !ok1 {
			break
		}
	}

	return true
}

// returns a receiver obly channel that will have DiffEvents
func DiffTrees(ctx context.Context, t1, t2 *tree.Tree) <-chan DiffEvent {
	appMetrics.mu.Lock()
	appMetrics.DiffTreeOpsTotal++
	appMetrics.mu.Unlock()

	diffs := make(chan DiffEvent)
	var wg sync.WaitGroup

	wg.Add(1)
	ch1 := make(chan int)
	go func() {
		defer wg.Done()
		Walk(ctx, t1, ch1)
		close(ch1)
	}()

	wg.Add(1)
	ch2 := make(chan int)
	go func() {
		defer wg.Done()
		Walk(ctx, t2, ch2)
		close(ch2)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(diffs)

		position := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("DiffTrees comparison cancelled: %v\n", ctx.Err())
				return
			default:

			}
			v1, ok1 := <-ch1
			v2, ok2 := <-ch2

			if !ok1 && !ok2 {
				fmt.Println("Both channels closed, no more values to compare.")
				return
			}
			position++
			if ok1 != ok2 {
				if ok1 {
					diffs <- DiffEvent{Type: "Missing Node T2", Value1: v1, Value2: 0, Position: position}
				} else {
					diffs <- DiffEvent{Type: "Missing Node T1", Value1: 0, Value2: v2, Position: position}
				}
				// diff found, increment the diffs found metric
				appMetrics.mu.Lock()
				appMetrics.DiffsFound++
				appMetrics.mu.Unlock()
				// keep going to get the rest of values for the channel open
				continue
			}
			if v1 != v2 {
				diffs <- DiffEvent{Type: "Different Values", Value1: v1, Value2: v2, Position: position}

				appMetrics.mu.Lock()
				appMetrics.DiffsFound++
				appMetrics.mu.Unlock()
			}
		}
	}()

	go func() {
		wg.Wait()
	}()

	return diffs

}

func main() {
	// Testing the Walk function with a context that can be cancelled
	// 5 is k val to create a tree with 10 nodes (5-50)
	fmt.Println("Single Tree Walk with time out to make sure it respects context cancellation")

	chTimed := make(chan int)
	ctxTimed, cancelTimed := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelTimed()

	go func() {
		Walk(ctxTimed, tree.New(5), chTimed)
		close(chTimed)
	}()

	go func() {
		for {
			select {
			case v, ok := <-chTimed:
				if !ok {
					fmt.Println("Channel closed, no more values to receive.")
					return
				}
				fmt.Printf("Received value: %d\n", v)
			case <-ctxTimed.Done():
				fmt.Println("Context timed out, stopping walk.")
				return
			}
		}
	}()
	time.Sleep(200 * time.Millisecond) // sleep to let the consumer go routine to react to cancellation
	// and enough time for main routine to finish concurrent execution before exiting
	// each Same func call starts up two goroutines
	fmt.Println("Same Tree Tests")
	fmt.Println("Same(tree.New(1), tree.New(1)) =", Same(tree.New(1), tree.New(1)))
	fmt.Println("Same(tree.New(1), tree.New(2)) =", Same(tree.New(1), tree.New(2)))
	fmt.Println("Same(tree.New(1), tree.New(1).Left) =", Same(tree.New(1), tree.New(1).Left))

	fmt.Println("Diff Trees Tests")
	treeA := tree.New(1)
	treeB := tree.New(1)
	treeC := tree.New(2)
	treeD := tree.New(1).Left

	ctxDiff1, cancelDiff1 := context.WithCancel(context.Background())
	defer cancelDiff1()
	diffChan1 := DiffTrees(ctxDiff1, treeA, treeB)
	for diff := range diffChan1 {
		fmt.Printf("Diff: Type: %s, Value1: %d, Value2: %d, Position: %d\n", diff.Type, diff.Value1, diff.Value2, diff.Position)
	}

	ctxDiff2, cancelDiff2 := context.WithCancel(context.Background())
	defer cancelDiff2()
	diffChan2 := DiffTrees(ctxDiff2, treeA, treeC)
	for diff := range diffChan2 {
		fmt.Printf("Diff: Type: %s, Value1: %d, Value2: %d, Position: %d\n", diff.Type, diff.Value1, diff.Value2, diff.Position)
	}

	ctxDiff3, cancelDiff3 := context.WithCancel(context.Background())
	defer cancelDiff3()
	diffChan3 := DiffTrees(ctxDiff3, treeA, treeD)
	for diff := range diffChan3 {
		fmt.Printf("Diff: Type: %s, Value1: %d, Value2: %d, Position: %d\n", diff.Type, diff.Value1, diff.Value2, diff.Position)
	}

	ctxDiffTimed, cancelDiffTimed := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelDiffTimed()
	diffChanTimed := DiffTrees(ctxDiffTimed, tree.New(10), tree.New(10))
	for diff := range diffChanTimed {
		fmt.Printf("Timed Diff: Type: %s, Value1: %d, Value2: %d, Position: %d\n", diff.Type, diff.Value1, diff.Value2, diff.Position)
	}

	time.Sleep(250 * time.Millisecond) // sleep to let the consumer go routine to react to cancellation
}
