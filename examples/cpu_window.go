package main

import (
	"log"
	"time"

	"github.com/kung-foo/go-metrics"
	"github.com/shirou/gopsutil/cpu"
)

func main() {
	// setup windows for a sample rate of 2Hz
	cpu5 := metrics.NewWindowSink(time.Second*5, 10)
	cpu15 := metrics.NewWindowSink(time.Second*15, 30)
	cpu60 := metrics.NewWindowSink(time.Second*60, 120)

	// use fanout sink to write to all three windows
	cpuSink := &metrics.FanoutSink{cpu5, cpu15, cpu60}

	key := []string{"cpu"}

	go func() {
		// collect CPU utilization every 500ms
		for _ = range time.Tick(time.Millisecond * 500) {
			v, _ := cpu.CPUPercent(0, false)
			cpuSink.AddSample(key, float32(v[0]))
		}
	}()

	// wait for first samples to come in
	time.Sleep(time.Second * 1)

	for _ = range time.Tick(time.Second * 1) {
		log.Printf("5s: %05.2f%%, 15s: %05.2f%%, 60s: %05.2f%%\n",
			cpu5.Sample(key).Mean(),
			cpu15.Sample(key).Mean(),
			cpu60.Sample(key).Mean())
	}
}
