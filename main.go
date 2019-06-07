package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
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

type heartBeatResultPerSec struct {
	Success    int
	Failed     int
	Duplicated int
}

func (rb *redisBench) run() error {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	rb.log(debug, "main start")

	var wg sync.WaitGroup
	before := MemConsumed()
	concurrentStream := make(chan interface{}, rb.Concurrent)
	heartBeatStream := make(chan HeartBeat, rb.RequestNum)
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
				heartBeatStream <- HeartBeat{Incr: incr.Val(), Time: now(), Status: Success}
				rb.logf(info, "%d", incr.Val())
			} else {
				heartBeatStream <- HeartBeat{Time: now(), Status: Failed}
				rb.logf(warn, "error %s", err)
			}
			rand.Seed(time.Now().UnixNano())
			sleepDuration := rb.Sleep + rand.Intn(rb.Sleep)
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			defer func() {
				wg.Done()
				client.Close()
				<-concurrentStream
			}()
		}()
	}
	wg.Wait()
	// channelはcloseしないとメモリリークの原因になる
	close(concurrentStream)
	close(heartBeatStream)
	incrRusult := map[int64]int{}
	heartBeatResult := map[int64]*heartBeatResultPerSec{}
	m := make(map[int64]bool)
	times := []int64{}
	for hb := range heartBeatStream {
		unixTime := hb.Time.Unix()
		if !m[unixTime] {
			m[unixTime] = true
			times = append(times, unixTime)
		}
		if _, ok := heartBeatResult[unixTime]; !ok {
			heartBeatResult[unixTime] = &heartBeatResultPerSec{
				Success:    0,
				Failed:     0,
				Duplicated: 0,
			}
		}
		if hb.Status == Success {
			heartBeatResult[unixTime].Success++
			if hb.Incr != 0 {
				if _, ok := incrRusult[hb.Incr]; ok {
					heartBeatResult[unixTime].Duplicated++
					rb.logf(warn, "duplicate!!! %d", hb.Incr)
				} else {
					incrRusult[hb.Incr] = 1
				}
			}
		} else {
			heartBeatResult[unixTime].Failed++
		}
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	for _, v := range times {
		fmt.Println(time.Unix(v, 0), heartBeatResult[v].Success, heartBeatResult[v].Failed, heartBeatResult[v].Duplicated)
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
