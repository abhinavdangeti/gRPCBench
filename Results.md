# Results

## Spec

```
MacBook Pro
Processor 2.9 GHz Intel Core i7
Memory 16 GB 2133 MHz LPDDR3
```

## Sample

```
2018-11-14 17:19:07.563044132 -0800 PST m=+60.002487855
Run time (secs): 60
Connection concurrency: 1
Message type: shipbulkdata
Num messages streamed per sample: 10

=======================================================
Client: client0
Addresses: [127.0.0.1:12345 127.0.0.1:23456]
=======================================================
10000 samples of 179560 events
Cumulative:	2.663875026s
HMean:		248.575µs
Avg.:		266.387µs
p50: 		237.313µs
p75:		268.012µs
p95:		481.747µs
p99:		621.806µs
p999:		761.495µs
Long 5%:	567.367µs
Short 5%:	190.234µs
Max:		1.002456ms
Min:		163.395µs
Range:		839.061µs
StdDev:		89.409µs
Rate/sec.:	3753.93
=======================================================
10000 samples of 179560 events
Cumulative:	2.804578049s
HMean:		263.574µs
Avg.:		280.457µs
p50: 		253.337µs
p75:		285.305µs
p95:		488.448µs
p99:		628.17µs
p999:		796.247µs
Long 5%:	578.632µs
Short 5%:	198.641µs
Max:		897.929µs
Min:		163.559µs
Range:		734.37µs
StdDev:		87.839µs
Rate/sec.:	3565.60
=======================================================

=======================================================
Client: client1
Addresses: [127.0.0.1:23456 127.0.0.1:34567]
=======================================================
10000 samples of 179560 events
Cumulative:	2.647154724s
HMean:		246.674µs
Avg.:		264.715µs
p50: 		234.429µs
p75:		265.785µs
p95:		483.521µs
p99:		619.817µs
p999:		765.571µs
Long 5%:	568.248µs
Short 5%:	190.383µs
Max:		876.209µs
Min:		168.733µs
Range:		707.476µs
StdDev:		90.109µs
Rate/sec.:	3777.64
=======================================================
10000 samples of 179560 events
Cumulative:	2.804823549s
HMean:		263.417µs
Avg.:		280.482µs
p50: 		253.073µs
p75:		284.125µs
p95:		496.039µs
p99:		628.748µs
p999:		784.594µs
Long 5%:	582.644µs
Short 5%:	200.083µs
Max:		1.034913ms
Min:		162.96µs
Range:		871.953µs
StdDev:		88.985µs
Rate/sec.:	3565.29
=======================================================

=======================================================
Client: client2
Addresses: [127.0.0.1:34567 127.0.0.1:12345]
=======================================================
10000 samples of 179560 events
Cumulative:	2.727125809s
HMean:		255.187µs
Avg.:		272.712µs
p50: 		243.12µs
p75:		273.711µs
p95:		493.612µs
p99:		632.993µs
p999:		744.357µs
Long 5%:	576.14µs
Short 5%:	196.732µs
Max:		910.557µs
Min:		166.804µs
Range:		743.753µs
StdDev:		89.415µs
Rate/sec.:	3666.86
=======================================================
10000 samples of 179560 events
Cumulative:	2.864990813s
HMean:		270.281µs
Avg.:		286.499µs
p50: 		261.803µs
p75:		293.311µs
p95:		490.84µs
p99:		628.752µs
p999:		782.886µs
Long 5%:	579.438µs
Short 5%:	199.856µs
Max:		978.119µs
Min:		172.807µs
Range:		805.312µs
StdDev:		86.031µs
Rate/sec.:	3490.41
=======================================================
```

## Comparison by varying connection concurrency

```
Run time (secs): 60
Message type: stream 10 msgs

+------------------------+-------------------+--------------+--------------+--------------+
| connection concurrency | samples collected |          avg |          p75 |          p95 |
+------------------------+-------------------+--------------+--------------+--------------+
|                      1 |            333298 |    276.268µs |    279.495µs |    491.044µs |
|                     10 |            593640 |   1.387657ms |   1.718451ms |   2.223423ms |
|                    100 |            647200 |  11.311667ms |  14.407776ms |  16.703479ms |
|                    500 |            657000 |  58.279957ms |  60.906887ms |  71.824427ms |
|                   1000 |            636000 | 126.885709ms | 133.873733ms | 151.025663ms |
|                   5000 |            480000 | 902.431268ms |  947.57333ms | 1.073416351s |
+------------------------+-------------------+--------------+--------------+--------------+
```

## Comparison of how ops/sec scales when client draws data from multiple servers

With a few changes to the client & server code, here are some results when 1 client is made to acquire data from 1 server and then 2 servers simultaneously (via multiple routines).
Note that the 2 servers are separate processes, however all the clients and the servers are on the same box.

```
Run time (secs): 30

+---------------------+---------------------+-------------------+-----------+
|      scenario       |  type of message(s) | samples collected | rate/sec. |
+---------------------+---------------------+-------------------+-----------+
|  1 client, 1 server |               hello |            223907 |   7099.49 |
| 1 client, 2 servers |               hello |            302858 |  11492.81 |
+---------------------+---------------------+-------------------+-----------+

+---------------------+---------------------+-------------------+-----------+
|      scenario       |  type of message(s) | samples collected | rate/sec. |
+---------------------+---------------------+-------------------+-----------+
|  1 client, 1 server |        stream 1 msg |            225957 |   7904.53 |
| 1 client, 2 servers |        stream 1 msg |            301574 |  11363.40 |
+---------------------+---------------------+-------------------+-----------+

+---------------------+---------------------+-------------------+-----------+
|      scenario       |  type of message(s) | samples collected | rate/sec. |
+---------------------+---------------------+-------------------+-----------+
|  1 client, 1 server |      stream 10 msgs |            184896 |   6327.68 |
| 1 client, 2 servers |      stream 10 msgs |            237736 |   8961.98 |
+---------------------+---------------------+-------------------+-----------+

+---------------------+---------------------+-------------------+-----------+
|      scenario       |  type of message(s) | samples collected | rate/sec. |
+---------------------+---------------------+-------------------+-----------+
|  1 client, 1 server | stream bulk 10 msgs |            181509 |   6219.76 |
| 1 client, 2 servers | stream bulk 10 msgs |            244962 |   9225.95 |
+---------------------+---------------------+-------------------+-----------+
```

## Comparison of results with varying stream lengths

With a few changes to the client & server code, here are some results when 1 client streams varying length responses from 2 servers simultaneously (via multiple routines).
Note that the 2 servers are separate processes, however all the clients and servers are on the same box.

```
Run time (secs): 30
Config: 1 client, 2 servers

+------------------------+--------------+-------------------+-------------+-------------+-------------+
| connection concurrency | num messages | samples collected |         avg |         p75 |         p95 |
+------------------------+--------------+-------------------+-------------+-------------+-------------+
|                      1 |           10 |            261254 |   200.493µs |   210.095µs |   268.185µs |
|                      1 |          100 |            103278 |   521.386µs |   568.697µs |   815.268µs |
|                      1 |        0-100 |            151594 |   332.375µs |    383.02µs |   597.004µs |
+------------------------+--------------+-------------------+-------------+-------------+-------------+
|                     10 |           10 |            786600 |   516.423µs |    583.12µs |   870.688µs |
|                     10 |          100 |            170540 |  2.644172ms |  3.078912ms |  3.611807ms |
|                     10 |        0-100 |            298200 |  1.387292ms |   1.70304ms |  2.084579ms |
+------------------------+--------------+-------------------+-------------+-------------+-------------+
|                    100 |           10 |            964000 |   3.49178ms |  4.282344ms |   5.00668ms |
|                    100 |          100 |            174400 | 24.493862ms | 30.090891ms | 33.405983ms |
|                    100 |        0-100 |            303800 | 12.470597ms | 15.465683ms | 18.838451ms |
+------------------------+--------------+-------------------+-------------+-------------+-------------+

```

## Comparison by varying server's MaxConcurrentStreams option

```
Run time (secs): 30
Config: 1 client, 2 servers
Connection concurrency: 1000
Num messages streamed per request: 100

+----------------------+-------------------+--------------+--------------+--------------+
| MaxConcurrentStreams | samples collected |          avg |          p75 |          p95 |
+----------------------+-------------------+--------------+--------------+--------------+
|              default |            136000 | 421.997874ms | 471.013867ms | 484.527011ms |
|                  500 |            158000 | 258.864737ms | 336.621661ms |  376.18497ms |
|                 1000 |            142000 |  372.38939ms | 408.299016ms | 425.797533ms |
|                 2000 |            142000 |   393.1159ms |  446.20276ms | 482.188021ms |
+----------------------+-------------------+--------------+--------------+--------------+

```

## Profiling

Flame graphs of CPU profiles on client that streams 10 messages from server(s)

* 1 client, 1 server, stream 10 msgs
![1client_1server](docs/flame_graph_1client_1server_shipdata10.png)

* 1 client, 2 servers, stream 10 msgs
![1client_2servers](docs/flame_graph_1client_2servers_shipdata10.png)
