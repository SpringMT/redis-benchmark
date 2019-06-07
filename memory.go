package main

import "runtime"

// MemConsumed 現在のメモリ消費量
func MemConsumed() uint64 {
	runtime.GC()
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return s.Sys
}
