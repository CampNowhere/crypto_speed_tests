package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
var startWaiter sync.Mutex

var threads = flag.Int("threads", 2, "Number of threads you'd like to run")
var hashes = flag.Int("hashes", 1000000, "Number of iterations you'd like to run")

func workerThread(iterations, threadID int) {
	buffer := make([]byte, 32)
	/*
		For synchronization purposes, we attempt to get a lock, the net effect being that we want to be able to
		control when all of the threads start. So we get a lock and then release it immediately, with the idea that
		each thread "waits", or blocks, until whoever holds the lock releases it. This allows us to start the dispatched
		threads nearly simultaneously while this lock/unlock cycles cascades through each thread.
	*/
	startWaiter.Lock()
	startWaiter.Unlock()
	for i := 0; i < iterations; i++ {
		out := sha256.Sum256(buffer)
		// Feeding the hash back into the algorithm, so we don't get any possible caching tricks speeding things up
		buffer = out[0:32]
	}
	fmt.Println("Thread", threadID, "coming home...")
	wg.Done()
}

func main() {
	flag.Parse()
	hashesPerThread := *hashes / *threads
	lastThreadHashes := hashesPerThread + (*hashes % *threads)
	// Locking our sync mutex so none of the threads start until we're ready
	startWaiter.Lock()
	var thisIterations int
	for i := 0; i < *threads; i++ {
		if i == (*threads-1) && lastThreadHashes > 0 {
			thisIterations = lastThreadHashes
		} else {
			thisIterations = hashesPerThread
		}
		go workerThread(thisIterations, i)
		fmt.Println("Dispatching thread", i, "with", thisIterations, "iterations...")
		// This is another thread synchronization primitive, essentially we're keeping track of how many threads
		// here and each thread will decrease this count when it's done.
		wg.Add(1)
	}
	t := time.Now()
	startWaiter.Unlock()
	// We block until each thread reports "done", otherwise the program will continue execution and exit
	wg.Wait()
	e := time.Now()
	dur := e.Sub(t)
	fmt.Println("Took", dur.Nanoseconds(), "nanoseconds")
	hps := (1000000000.0 / float32(dur.Nanoseconds())) * float32(*hashes)
	fmt.Printf("System calculated %.2f million SHA256 hashes per second!\n", hps/1000000.0)
}
