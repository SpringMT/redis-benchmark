package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
)

type redisBench struct {
	Host                 string `long:"host" description:"redisのhost名"`
	Concurrent           int    `short:"c" default:"50" description:"並列接続数"`
	RequestNum           int    `short:"n" default:"100" description:"リクエスト総数"`
	Sleep                int    `short:"s" default:"1000" description:"スリープ ミリ秒"`
	Verbose              []bool `short:"v" long:"verbose" description:"verbose output. it can be stacked like -vv for more detailed log"`
	outStream, errStream io.Writer
}

func (rb *redisBench) run() error {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	rb.log(debug, "main start")

	var wg sync.WaitGroup
	before := MemConsumed()
	concurrentStream := make(chan interface{}, rb.Concurrent)
	heartBeatStream := make(chan heartBeat, rb.RequestNum)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT)
	for i := 0; i < rb.RequestNum; i++ {
		wg.Add(1)
		concurrentStream <- true
		go func() {
			rb.log(debug, "redis incr start")
			client := RedisNewClient(rb.Host)
			res, err := client.increment("pipeline_counter")
			if err == nil {
				heartBeatStream <- heartBeat{Incr: res, Time: now(), Status: Success}
				rb.logf(info, "%d", res)
			} else {
				heartBeatStream <- heartBeat{Time: now(), Status: Failed}
				rb.logf(warn, "error %s", err)
			}
			time.Sleep(time.Duration(rb.Sleep+rand.Intn(rb.Sleep)) * time.Millisecond)
			defer func() {
				wg.Done()
				client.close()
				<-concurrentStream
			}()
		}()
	}
	wg.Wait()
	after := MemConsumed()
	rb.logf(info, "Memory %.3f kb", float64(after-before)/1000)
	// channelはcloseしないとメモリリークの原因になる
	close(concurrentStream)
	close(heartBeatStream)
	results := NewHeartBeatResult()
	for hb := range heartBeatStream {
		unixTime := hb.Time.Unix()
		results.add(unixTime, hb.Incr, hb.Status)
	}
	results.show()
	return nil
}

func now() *time.Time {
	now := time.Now()
	return &now
}

func parseArgs(args []string) (*flags.Parser, *redisBench, error) {
	rb := &redisBench{}
	p := flags.NewParser(rb, flags.Default)
	p.Usage = fmt.Sprintf(`--host localhost [...]

Version: %s (rev: %s/%s)`, version, revision, runtime.Version())
	_, err := p.ParseArgs(args)
	rb.outStream = os.Stdout
	rb.errStream = os.Stderr
	return p, rb, err
}

// Run benchmark
func Run(args []string) int {
	p, rb, err := parseArgs(args)
	if err != nil {
		if ferr, ok := err.(*flags.Error); !ok || ferr.Type != flags.ErrHelp {
			p.WriteHelp(rb.errStream)
		}
		return 2
	}
	runErr := rb.run()
	if runErr != nil {
		return wrapcommander.ResolveExitCode(runErr)
	}

	return 0
}

func main() {
	rand.Seed(time.Now().UnixNano())
	os.Exit(Run(os.Args[1:]))
}
