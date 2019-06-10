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
	Time     *time.Time
	Incr     int64
	Status   HeartBeatStatus
	Duration time.Duration
}

type heartBeatResultPerSec struct {
	Success   int
	Failed    int
	Incrs     []int64
	Durations []time.Duration
}

type heartBeatResults struct {
	Results map[int64]*heartBeatResultPerSec
}

func (result *heartBeatResults) add(hb heartBeat) {
	key := hb.Time.Unix()
	if _, ok := result.Results[key]; !ok {
		result.Results[key] = &heartBeatResultPerSec{
			Success:   0,
			Failed:    0,
			Incrs:     []int64{},
			Durations: []time.Duration{},
		}
	}
	if hb.Status == Success {
		result.Results[key].Success++
		result.Results[key].Incrs = append(result.Results[key].Incrs, hb.Incr)
		result.Results[key].Durations = append(result.Results[key].Durations, hb.Duration)
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
		min, max, ave := calcMinAndMaxAndAve(result.Results[v].Durations)
		fmt.Println(time.Unix(v, 0), res.Success, res.Failed, duplicated, fmt.Sprintf("%.2f ms", min), fmt.Sprintf("%.2f ms", max), fmt.Sprintf("%.2f ms", ave))
	}
	for k, v := range duplicatedResult {
		fmt.Printf("Incremented Value %d is duplicated at %s\n", k, strings.Join(v, ","))
	}
}

func calcMinAndMaxAndAve(durations []time.Duration) (minMillSec float64, maxMillSec float64, aveMillSec float64) {
	if len(durations) == 0 {
		return 0, 0, 0
	}
	minDuration := durations[0]
	maxDuration := durations[0]
	var sumDuration time.Duration
	count := 0
	for _, duration := range durations {
		count++
		sumDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
		if duration < minDuration {
			minDuration = duration
		}
	}
	min := float64(minDuration / time.Millisecond)
	max := float64(maxDuration / time.Millisecond)
	ave := float64(float64(sumDuration/time.Millisecond) / float64(count))
	return min, max, ave
}

func NewHeartBeatResult() heartBeatResults {
	return heartBeatResults{Results: map[int64]*heartBeatResultPerSec{}}
}
