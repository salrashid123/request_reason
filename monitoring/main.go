package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"sync"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	wg sync.WaitGroup

	mCount     = stats.Int64("Count", "# number of called..", stats.UnitNone)
	keyPath, _ = tag.NewKey("customerID")
	countView  = &view.View{
		Name:        "gcs/requests",
		Measure:     mCount,
		Description: "The count of calls per customer",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{keyPath},
	}
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
	reasons := make([]string, 0)
	reasons = append(reasons,
		"customer-G",
		"customer-A",
		"customer-T",
		"customer-C",
	)
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "app",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus exporter: %v", err)
	}
	err = view.Register(countView)
	if err != nil {
		log.Fatalf("Could not register view %v", err)
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("Failed to run Prometheus /metrics endpoint: %v", err)
		}
	}()

	ctx := context.Background()

	// simulate random customer traffic
	for i := 1; i < 15000; i++ {
		reason := reasons[rand.Intn(len(reasons))]
		wg.Add(1)
		go func(ctx context.Context, reason string) {
			defer wg.Done()
			n := rand.Intn(500)
			time.Sleep(time.Duration(n) * time.Second)
			ctx, err := tag.New(context.Background(), tag.Insert(keyPath, reason))
			if err != nil {
				log.Println(err)
			}
			stats.Record(ctx, mCount.M(1))
			log.Printf("Emitting Metric for %s", reason)
		}(ctx, reason)
	}

	// wait 3 minutes and then send a burst of additional traffic for one customer
	time.Sleep(3 * time.Minute)
	reason := "customer-G"
	for i := 1; i < 2000; i++ {
		wg.Add(1)
		go func(ctx context.Context, reason string) {
			defer wg.Done()
			n := rand.Intn(100)
			time.Sleep(time.Duration(n) * time.Second)
			ctx, err := tag.New(context.Background(), tag.Insert(keyPath, reason))
			if err != nil {
				log.Println(err)
			}
			stats.Record(ctx, mCount.M(1))
			log.Printf("Emitting Metric for %s", reason)
		}(ctx, reason)
	}

	wg.Wait()

}
