package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type HeartBeatStatus int

const (
	Success HeartBeatStatus = iota
	Failed
)

type heartBeat struct {
	Time   *time.Time
	Incr   int64
	Status HeartBeatStatus
}

type heartBeatResultPerSec struct {
	Success int
	Failed  int
	Incrs   []int64
}

type heartBeatResults struct {
	Results map[int64]*heartBeatResultPerSec
}

func (result *heartBeatResults) add(key int64, incr int64, status HeartBeatStatus) {
	if _, ok := result.Results[key]; !ok {
		result.Results[key] = &heartBeatResultPerSec{
			Success: 0,
			Failed:  0,
			Incrs:   []int64{},
		}
	}
	if status == Success {
		result.Results[key].Success++
		result.Results[key].Incrs = append(result.Results[key].Incrs, incr)
	} else {
		result.Results[key].Failed++
	}
}

func (result *heartBeatResults) show() {
	times := []int64{}
	for k := range result.Results {
		times = append(times, k)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	incrRusult := map[int64]int{}
	duplicatedResult := map[int64][]string{}
	for _, v := range times {
		res := result.Results[v]
		duplicated := 0
		for _, incr := range res.Incrs {
			if _, ok := incrRusult[incr]; ok {
				duplicated++
				duplicatedResult[incr] = append(duplicatedResult[incr], time.Unix(v, 0).Format("2006/01/02 15:04:05"))
			} else {
				incrRusult[incr] = 1
			}
		}
		fmt.Println(time.Unix(v, 0), res.Success, res.Failed, duplicated)
	}
	for k, v := range duplicatedResult {
		fmt.Printf("Incremented Valu %d is duplicated at %s\n", k, strings.Join(v, ","))
	}
}

func NewHeartBeatResult() heartBeatResults {
	return heartBeatResults{Results: map[int64]*heartBeatResultPerSec{}}
}
