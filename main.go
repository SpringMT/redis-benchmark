package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"sync"
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

type HeartBeatStatus int

const (
	Success HeartBeatStatus = iota
	Failed
)

type HeartBeat struct {
	Incr   int64
	Time   *time.Time
	Status HeartBeatStatus
}

func (rb *redisBench) run() error {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	rb.log(debug, "main start")

	var wg sync.WaitGroup
	before := MemConsumed()
	concurrentStream := make(chan bool, rb.Concurrent)
	haertBeatStream := make(chan HeartBeat, rb.RequestNum)
	for i := 0; i < rb.RequestNum; i++ {
		wg.Add(1)
		concurrentStream <- true
		go func() {
			rb.log(debug, "redis incr start")
			client := RedisNewClient(rb.Host)
			pipe := client.Pipeline()
			incr := pipe.Incr("pipeline_counter")
			_, err := pipe.Exec()
			if err == nil {
				haertBeatStream <- HeartBeat{Incr: incr.Val(), Time: now(), Status: Success}
				rb.logf(info, "%d", incr.Val())
			} else {
				haertBeatStream <- HeartBeat{Incr: -1, Time: now(), Status: Failed}
				rb.logf(warn, "error %s", err)
			}
			time.Sleep(time.Duration(rb.Sleep) * time.Millisecond)
			defer func() {
				wg.Done()
				if client != nil {
					client.Close()
				}
				<-concurrentStream
			}()
		}()
	}
	wg.Wait()
	close(haertBeatStream)
	incrRusult := map[int64]int{}
	heartBeatResult := map[int64]map[string]int{}
	m := make(map[int64]bool)
	times := []int64{}
	for hb := range haertBeatStream {
		unixTime := hb.Time.Unix()
		if !m[unixTime] {
			m[unixTime] = true
			times = append(times, unixTime)
		}
		if _, ok := heartBeatResult[unixTime]; !ok {
			heartBeatResult[unixTime] = map[string]int{"success": 0, "failed": 0, "duplicated": 0}
		}
		if hb.Status == Success {
			heartBeatResult[unixTime]["success"]++
		} else {
			heartBeatResult[unixTime]["failed"]++
		}

		if hb.Incr != -1 {
			if _, ok := incrRusult[hb.Incr]; ok {
				heartBeatResult[unixTime]["duplicated"]++
				rb.logf(info, "duplicate!!! %d", hb.Incr)
			} else {
				incrRusult[hb.Incr] = 1
			}
		}
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	for _, v := range times {
		fmt.Println(time.Unix(v, 0), heartBeatResult[v]["success"], heartBeatResult[v]["failed"], heartBeatResult[v]["duplicated"])
	}
	after := MemConsumed()
	rb.logf(info, "%.3f kb", float64(after-before)/1000)
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
	os.Exit(Run(os.Args[1:]))
}
