caffy-bgp
===

Caffy-bgp is a simple BGP client that will take in a BGP feed and any updates that come form the peer,
will be sent into a redis pubsub topic so that it can be consumed by other applications easily.

Example JSON events currently look like this:

```
{
  "Message": null,
  "PeerAS": 60011,
  "LocalAS": 4242421338,
  "PeerAddress": "216.66.80.190",
  "LocalAddress": "188.165.140.182",
  "PeerID": "185.101.96.2",
  "FourBytesAs": true,
  "Timestamp": "2016-12-22T14:05:32.476927786Z",
  "Payload": "/////////////////////wBqAgAAAEtAAQEAQAIuAgsAAOprAACqCwAADRwAAA0cAAANHAAADRwAAA0cAAANHAAAGTUAABKTAACy/EADBNhCUL7ACAzqa/pv+m8MMPpvqgsYPQxfGD0MLg==",
  "PostPolicy": true,
  "PathList": [
    {
      "nlri": {
        "prefix": "61.12.95.0/24"
      },
      "attrs": [
        {
          "type": 1,
          "value": 0
        },
        {
          "type": 2,
          "as_paths": [
            {
              "segment_type": 2,
              "num": 11,
              "asns": [
                60011,
                43531,
                3356,
                3356,
                3356,
                3356,
                3356,
                3356,
                6453,
                4755,
                45820
              ]
            }
          ]
        },
        {
          "type": 3,
          "nexthop": "216.66.80.190"
        },
        {
          "type": 8,
          "communities": [
            3932945007,
            4201581616,
            4201622027
          ]
        }
      ],
      "age": 1482415532,
      "source-id": "185.101.96.2",
      "neighbor-ip": "216.66.80.190"
    },
    {
      "nlri": {
        "prefix": "61.12.46.0/24"
      },
      "attrs": [
        {
          "type": 1,
          "value": 0
        },
        {
          "type": 2,
          "as_paths": [
            {
              "segment_type": 2,
              "num": 11,
              "asns": [
                60011,
                43531,
                3356,
                3356,
                3356,
                3356,
                3356,
                3356,
                6453,
                4755,
                45820
              ]
            }
          ]
        },
        {
          "type": 3,
          "nexthop": "216.66.80.190"
        },
        {
          "type": 8,
          "communities": [
            3932945007,
            4201581616,
            4201622027
          ]
        }
      ],
      "age": 1482415532,
      "source-id": "185.101.96.2",
      "neighbor-ip": "216.66.80.190"
    }
  ]
}
```

To configure a BGP session, You will need to point it a JSON config file in the following layout:

```
[
  {
    "peeraddr": "192.168.2.1",
    "peeras": 4242421337,
    "localaddr": "192.168.2.50"
  }
]
```

Where 4242421337 is the ASN you are peering with, the ASN that this program uses is set though the
`-myasn 1234` command line option.

Usage:

```
$ ./caffy-bgp -h
Usage of ./caffy-bgp:
  -bgpport int
    the port you want to run bgp on (default 179)
  -cfgfile string
    a json array of bgp peers (default "./peers.json")
  -grpc
    enable grpc/gobgp commandage (default true)
  -myasn int
    The ASN of this running program (default 4242421338)
  -redis string
    redis address (default "localhost:6379")
  -redis-topic string
    the redis pubsub topic (default "bgp-caffy")
  -routerid string
    the bgp router id (default "192.168.2.50")
  -statsbind string
    http stats bind (default "127.0.0.1:56565")
```

You can monitor the health of the program using the metrics endpoint defined by `-statsbind`
and by default is set to: `127.0.0.1:56565/debug/vars`

