## Example of an issue with executing Osquery over Go client

### Steps to reproduce

1. Start ```osqueryd -S``` shell if osquery is not already running
2. Find osquery socket path 
```
osqueryi --nodisable_extensions
osquery> select value from osquery_flags where name = 'extensions_socket';
+-----------------------------------+
| value                             |
+-----------------------------------+
| /Users/USERNAME/.osquery/shell.em |
+-----------------------------------+
```
3. Start local http server example, it allows to emulate the long running queries over osquery ```curl``` table.
```
go run server/main.go
```

4. Execute the sample code from this repo that shows the problem
```
sudo go run main.go --socket /Users/USERNAME/.osquery/shell.em
```

The timeout for osquery client constructor is passed as 10 secs in this case.

Here is what the example does:
1. Executes the query #1 that should take 30s to run
2. Executes the query #2 that should take 5s to run
3. After completion of steps above it executes the query #3 that should return right away until it succeeds.

As the result:
1. While query #1 is executed, the query #2 fails right away, which is expected since the locker default wait time is 200ms
https://github.com/osquery/osquery-go/blob/master/client.go#L16
2. Query #1 times out in 10 secs, which is also kind of expected since the client was initialized with 10s timeout.
3. Here where it gets interesting. The query #3, that should return right away
    a. First attempt sails with ```query: out of order sequence response```
    b. Second attempt fails with ```i/o timeout``` after 10s timeout
    c. And then on third attempt is succeeds


Here is the output:
```
[213.291µs] Execute query: select * from curl where url='http://localhost:8080/?sleep=30s'
[500.556833ms] Execute query: select * from curl where url='http://localhost:8080/?sleep=5s'
[705.269333ms] Failed query: select * from curl where url='http://localhost:8080/?sleep=5s', err: timeout after 200ms
[10.001679916s] Failed query: select * from curl where url='http://localhost:8080/?sleep=30s', err: read unix ->/Users/[redacted]/.osquery/shell.em: i/o timeout
[10.00192125s] Execute query: select * from curl where url='http://localhost:8080/?sleep=0s'
[16.005457916s] Failed query:  select * from curl where url='http://localhost:8080/?sleep=0s', err: query: out of order sequence response
[16.005506083s] Execute query: select * from curl where url='http://localhost:8080/?sleep=0s'
[26.033969291s] Failed query:  select * from curl where url='http://localhost:8080/?sleep=0s', err: read unix ->/Users/[redacted]/.osquery/shell.em: i/o timeout
[26.034065958s] Execute query: select * from curl where url='http://localhost:8080/?sleep=0s'
[map[bytes:5 method:GET response_code:200 result:Done
 round_trip_time:5304 url:http://localhost:8080/?sleep=0s user_agent:osquery]]
```


For now it seems like due this particular issues we would have to capture and possibly retry the queries that timeout.