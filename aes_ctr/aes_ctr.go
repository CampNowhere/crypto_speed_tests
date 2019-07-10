package main

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"sync"
	"time"
)

type initializationVector []byte

var keyHex = "5826d6ae4e378e26c6dd6c14046d624d7cb168d30517cb57506d5518f29ab5c3"
var ivHex = "242bd8fb399de07a0000000000000000"
var blockSize = aes.BlockSize

// AES block size is 16, the following number yields a gig of ram
var numberOfBlocks = 67108864

//var numberOfBlocks = 20
var waiter sync.Mutex
var wg sync.WaitGroup

var iv, key, pt []byte

var threads = flag.Int("threads", 2, "set to the number of threads you want to use to concurrently encrypt")

func (iv initializationVector) ctrSetBlockCount(c uint64) {
	binary.LittleEndian.PutUint64(iv[8:16], c)
}

func encryptThread(min, max, threadID int) {
	fmt.Printf("Starting thread %v with blocks %v through %v\n", threadID, min, max-1)
	crypter, _ := aes.NewCipher(key)
	myIv := make([]byte, 16)
	buf := make([]byte, 16)
	var blockAddress int
	copy(myIv, iv)
	waiter.Lock()
	waiter.Unlock()
	for i := min; i < max; i++ {
		initializationVector(myIv).ctrSetBlockCount(uint64(i))
		blockAddress = i * blockSize
		crypter.Encrypt(buf, myIv)
		sliceXor(pt[blockAddress:blockAddress+blockSize], buf)
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
	pt = make([]byte, blockSize*numberOfBlocks)
	blocksPerThread := numberOfBlocks / *threads
	currentBlock := 0
	waiter.Lock()
	for i := 0; i < *threads; i++ {
		min := currentBlock
		max := currentBlock + blocksPerThread
		if i == (*threads - 1) {
			max += numberOfBlocks % *threads
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
	fmt.Printf("Encrypted %v blocks in %v nanoseconds with %v threads\n", numberOfBlocks, elapsed.Nanoseconds(), *threads)
}
