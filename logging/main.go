package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	CLOUDSDK_CORE_REQUEST_REASON_ENV_VAR = "CLOUDSDK_CORE_REQUEST_REASON"
)

var (
	projectID = "mineral-minutia-820"
	bucket    = "mineral-minutia-820-bucket"
	wg        sync.WaitGroup
)

func main() {
	flag.Parse()

	ctx := context.Background()

	rand.Seed(time.Now().Unix())
	reasons := make([]string, 0)
	reasons = append(reasons,
		"customer-A",
		"customer-B",
		"customer-C",
		"customer-D",
		"customer-E",
		"customer-F",
	)

	for i := 1; i < 20; i++ {
		company := reasons[rand.Intn(len(reasons))]

		wg.Add(1)
		go func(ctx context.Context, reason string) {
			defer wg.Done()

			client, err := storage.NewClient(ctx, option.WithRequestReason(reason))
			if err != nil {
				fmt.Printf("Error creating Client: %v", err)
				return
			}

			attr, err := client.Bucket(bucket).Attrs(ctx)
			if err != nil {
				fmt.Printf("Error Reading bucket attrs: %v", err)
				return
			}
			fmt.Printf("BucketName: %s\n", attr.Name)
		}(ctx, company)

	}

	wg.Wait()
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
