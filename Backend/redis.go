package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis-11588.c212.ap-south-1-1.ec2.redns.redis-cloud.com:11588",
		Username: "default",
		Password: "L1upv40jP88QWBFsUS70ckUNPsw0veRY",
		DB:       0,
	})

	// Measure Redis round-trip time (network + processing)
	const numPings = 5
	var latencies []time.Duration

	fmt.Printf("Performing %d ping tests...\n", numPings)

	for i := 0; i < numPings; i++ {
		start := time.Now()
		_, err := rdb.Ping(ctx).Result()
		latency := time.Since(start)

		if err != nil {
			panic(err)
		}

		latencies = append(latencies, latency)
		fmt.Printf("Ping %d: %v (%.2f ms)\n", i+1, latency, float64(latency.Nanoseconds())/1000000)
	}

	// Calculate statistics
	var total time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(len(latencies))

	fmt.Printf("\nRound-trip time statistics:\n")
	fmt.Printf("Average: %v (%.2f ms)\n", avg, float64(avg.Nanoseconds())/1000000)
	fmt.Printf("Min: %v (%.2f ms)\n", min, float64(min.Nanoseconds())/1000000)
	fmt.Printf("Max: %v (%.2f ms)\n", max, float64(max.Nanoseconds())/1000000)
	fmt.Printf("\nNote: This measures total round-trip time including network latency and minimal Redis processing time.\n")
}

