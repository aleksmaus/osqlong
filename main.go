package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/osquery/osquery-go"
)

// This the url to the server that you can run from this project
// go run server/main.go
const httpUrlBase = "http://localhost:8080"

var socketTimeout = 10 * time.Second

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func renderQuery(sleep time.Duration) string {
	return fmt.Sprintf("select * from curl where url='http://localhost:8080/?sleep=%v'", sleep)
}

// Execute: go run main.go --socket /Users/USERNAME/.osquery/shell.em
// Use the osquery socket path on your platform
//
// osqueryi --nodisable_extensions
// osquery> select value from osquery_flags where name = 'extensions_socket';
// +-----------------------------------+
// | value                             |
// +-----------------------------------+
// | /Users/USERNAME/.osquery/shell.em |
// +-----------------------------------+
func main() {
	socket := flag.String("socket", "", "Path to osquery socket file")
	flag.Parse()
	if *socket == "" {
		log.Fatalf(`Usage: %s --socket SOCKET_PATH`, os.Args[0])
	}

	ctx := context.Background()

	cli, err := osquery.NewClient(*socket, socketTimeout)
	exitOnError(err)
	defer cli.Close()

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		query := renderQuery(30 * time.Second)
		fmt.Printf("[%v] Execute query: %v\n", time.Since(start), query)
		res, err := cli.QueryContext(ctx, query)
		if err != nil {
			fmt.Printf("[%v] Failed query: %s, err: %v\n", time.Since(start), query, err)
			return
		}
		fmt.Println(res.Response)
	}()

	go func() {
		defer wg.Done()
		// Delay this query to make sure it is executed after the first query is in flight
		time.Sleep(500 * time.Millisecond)
		query := renderQuery(5 * time.Second)
		fmt.Printf("[%v] Execute query: %v\n", time.Since(start), query)
		res, err := cli.QueryContext(ctx, query)
		if err != nil {
			fmt.Printf("[%v] Failed query: %s, err: %v\n", time.Since(start), query, err)
			return
		}
		fmt.Println(res.Response)
	}()

	wg.Wait()

	// Try to run query after timeout
	for {
		query := renderQuery(time.Duration(0))
		fmt.Printf("[%v] Execute query: %v\n", time.Since(start), query)
		res, err := cli.QueryContext(ctx, query)
		if err != nil {
			fmt.Printf("[%v] Failed query:  %s, err: %v\n", time.Since(start), query, err)
			//return
		} else {
			fmt.Println(res.Response)
			return
		}
	}
}
