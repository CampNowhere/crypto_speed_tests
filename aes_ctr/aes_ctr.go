package main

import (
	"crypto/aes"
	"encoding/hex"
	"flag"
	"fmt"
	"sync"
	"time"
)

var keyHex = "5826d6ae4e378e26c6dd6c14046d624d7cb168d30517cb57506d5518f29ab5c3"
var ivHex = "242bd8fb399de07a0000000000000000"
var blockSize = aes.BlockSize

var waiter sync.Mutex
var wg sync.WaitGroup

var iv, key []byte

// AES block size is 16, the default number yields a gig of ram
var numberOfBlocks = flag.Int("blocks", 67108864, "number of AES blocks you want to encrypt")
var threads = flag.Int("threads", 2, "set to the number of threads you want to use to concurrently encrypt")

func encryptThread(min, max, threadID int) {
	fmt.Printf("Starting thread %v with blocks %v through %v\n", threadID, min, max-1)
	crypter, _ := aes.NewCipher(key)
	myIv := make([]byte, 16)
	buf := make([]byte, 16)
	copy(myIv, iv)
	// Starting each thread with a unique IV so each thread isn't doing the exact same work
	myIv[len(myIv)-1] = byte(threadID)
	waiter.Lock()
	waiter.Unlock()
	for i := min; i < max; i++ {
		crypter.Encrypt(buf, myIv)
		copy(myIv, buf)
	}
	wg.Done()
}

func sliceXor(dst, src []byte) {
	for i := range dst {
		dst[i] = dst[i] ^ src[i]
	}
}

func main() {
	flag.Parse()
	key, _ = hex.DecodeString(keyHex)
	iv, _ = hex.DecodeString(ivHex)
	blocksPerThread := *numberOfBlocks / *threads
	currentBlock := 0
	waiter.Lock()
	for i := 0; i < *threads; i++ {
		min := currentBlock
		max := currentBlock + blocksPerThread
		if i == (*threads - 1) {
			max += *numberOfBlocks % *threads
		}
		go encryptThread(min, max, i)
		currentBlock += blocksPerThread
		wg.Add(1)
	}
	t1 := time.Now()
	waiter.Unlock()
	wg.Wait()
	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Printf("Encrypted %v blocks in %v nanoseconds with %v threads\n", *numberOfBlocks, elapsed.Nanoseconds(), *threads)
	blocksPerSecond := 1000000000.0 / float32(elapsed.Nanoseconds()) * float32(*numberOfBlocks)
	fmt.Printf("Blocks calculated at a rate of %.2f million per second with %v threads\n", blocksPerSecond/1000000.0, *threads)
}
