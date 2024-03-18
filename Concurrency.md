Concurrency

- [**wg.Add**](#wgadd)
- [**TryLock**](#trylock)
- [**sync.RWMutex**](#syncrwmutex)
- [**sync.Cond**](#synccond)
- [**Barrier pattern**](#barrier-pattern)
- [**Select with default computation**](#select-with-default-computation)
- [**Fan out, fan in**](#fan-out-fan-in)
- [**Broadcast**](#broadcast)
- [**Prime number**](#prime-number)
- [**Worker Pool Semaphore pattern**](#worker-pool-semaphore-pattern)
- [**Timer**](#timer)


## [**wg.Add**](https://www.storj.io/blog/production-concurrency)


```go
func processConcurrently(item []*Item) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, item := range items {
		item := item
		go func() {
			process(&wg, item)
		}()
	}
}

func process(wg *sync.WaitGroup, item *Item) {
	wg.Add(1)
	defer wg.Done()

	...
}
```

In this case `processConcurrently` can exit before `wg.Add` is called

Call `wg.Add` right before spawning the goroutine:

```go
var wg sync.WaitGroup
defer wg.Wait()
...
for ... {
	wg.Add(1)
	go func() {
		defer wg.Done()
```


## **[TryLock](https://pkg.go.dev/sync#Mutex.TryLock)**

An example of a possible use of `TryLock` is for a monitoring goroutine: if `TryLock` returns false (already locked) then it can decide to try again later


## **[sync.RWMutex](https://pkg.go.dev/sync#RWMutex)**

More efficient to allow multiple readers to access the shared critical area
* `RLock` (reader lock) prevents other goroutine from acquiring the writing lock but does not block other goroutine from also acquiring the reader lock


## **[sync.Cond](https://pkg.go.dev/sync#Cond)**

`Cond.Signal` will wake up a goroutine that called `Cond.Wait` but if there are no goroutine waiting then the signal gets lost. No control over which sleeping goroutine is chosen.

`Cond.Broadcast` will wake up all sleeping goroutines that called `Cond.Wait`, synchronizing them all together.

```go
func playerHandler(cond *sync.Cond, playersRemaining *int, playerId int) {
	cond.L.Lock()
	fmt.Println(playerId, ": Connected")
	*playersRemaining--
	if *playersRemaining == 0 {
		cond.Broadcast()
	}
	for *playersRemaining > 0 {
		fmt.Println(playerId, ": Waiting for more players")
		cond.Wait()
	}
	cond.L.Unlock()
	fmt.Println("All players connected. Ready player", playerId)
	//Game started
}
```

Always use Signal(), Broadcast() and Wait() while holding the mutex lock to avoid synchronization problems

```go
cond.L.Lock()
cond.Signal()
cond.L.Unlock()
```

## **Barrier pattern**

* https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter6/listing6.10/barrier.go
* https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter6/listing6.16_17/matrixmultiply.go

Synchronize start and end. For loop to avoid new goroutine spawnings (not really that beneficial since goroutines are lightweight).

Barrier with a `waitCount` of `n + 1` so that you can decide when to call `barrier.Wait` to synchronize all the goroutines.


## **Select with default computation**

* https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter8/listing8.4_5_6/passwordguesser.go#L31

Use the `default` branch of a `select` block to perform a computation while the `case` block checks for a closing signal.


## **Fan out, fan in**

* Fan out: multiple goroutines read from the same channel

```go
// https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter9/listing9.9_11/extractwordsmulti.go
func downloadPages(quit <-chan int, urls <-chan string) <-chan string
func generateUrls(quit <-chan int) <-chan string

// Both downloadPages and generateUrls spawn a separate goroutine

func main(){
 quit := make(chan int)
    defer close(quit)
    urls := generateUrls(quit)
    pages := make([]<-chan string, downloaders)
    for i := 0; i < downloaders; i++ {
        pages[i] = downloadPages(quit, urls)
    }
	//
}
```

* Fan in: merge contents from multiple channels

```go
// https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter9/listing9.10/fanIn.go
func FanIn[K any](quit <-chan int, allChannels ...<-chan K) chan K {
    wg := sync.WaitGroup{}
    wg.Add(len(allChannels))
    output := make(chan K)
    for _, c := range allChannels {
        go func(channel <-chan K) {
            defer wg.Done()
            for i := range channel {
                select {
                case output <- i:
                case <-quit:
                    return
                }
            }
        }(c)
    }
    go func() {
        wg.Wait()
        close(output)
    }()
    return output
}
```



## **Broadcast**

Replicate the content to a set of output channels

```go
// https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter9/listing9.14/broadcast.go

func Broadcast[K any](quit <-chan int, input <-chan K, n int) []chan K {
    outputs := CreateAll[K](n)
    go func() {
        defer CloseAll(outputs...)
        var msg K
        moreData := true
        for moreData {
            select {
            case msg, moreData = <-input:
                if moreData {
                    for _, output := range outputs {
                        output <- msg
                    }
                }
            case <-quit:
                return
            }
        }
    }()
    return outputs
}
```


## **Prime number**

https://github.com/cutajarj/ConcurrentProgrammingWithGo/blob/main/chapter9/listing9.20_21/primesieve.go



## [**Worker Pool Semaphore pattern**](https://www.youtube.com/watch?v=5zXAHh5tJqQ&t=31m30s)

Only one goroutine is blocked at time, which is the one waiting for the signal on the semaphore channel:

```go
func main() {
	var limit = 2
	hugeSlice := []string{
		"task 1",
		"task 2",
		"task 3",
		"task 4",
	}
	sem := make(chan struct{}, limit)
	for _, task := range hugeSlice {
		// Acquire token
		sem <- struct{}{}
		go func(task string) {
			// perform task
			fmt.Println(task)
			time.Sleep(time.Second)
			// Release token
			<-sem
		}(task)
	}
	// Wait until all tokens are released
	for n := limit; n > 0; n-- {
		sem <- struct{}{}
	}
}
```


## **Timer**

https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers/