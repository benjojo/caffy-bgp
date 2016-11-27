package main

import (
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/config"
	gobgp "github.com/osrg/gobgp/server"
	"gopkg.in/redis.v3"
)

var (
	statbindaddr = flag.String("statsbind", "127.0.0.1:56565", "http stats bind")
	configfile   = flag.String("cfgfile", "./peers.json", "a json array of bgp peers")
	routerid     = flag.String("routerid", "192.168.2.50", "the bgp router id")
	selfasn      = flag.Int("myasn", 4242421338, "The ASN of this running program")
	bgpport      = flag.Int("bgpport", 179, "the port you want to run bgp on")
	enablegrpc   = flag.Bool("grpc", true, "enable grpc/gobgp commandage")
	redisaddr    = flag.String("redis", "localhost:6379", "redis address")
	redistopic   = flag.String("redis-topic", "bgp-caffy", "the redis pubsub topic")
	PublishChan  chan string
)

var (
	bgpUpdates   *expvar.Int
	bgpWithdraws *expvar.Int
)

type BGPPeer struct {
	Localaddr string `json:"localaddr"`
	Peeraddr  string `json:"peeraddr"`
	Peeras    int    `json:"peeras"`
}

func main() {
	fmt.Printf("1")
	bgpUpdates = expvar.NewInt("Updates")
	bgpWithdraws = expvar.NewInt("Withdraws")
	flag.Parse()

	PublishChan = make(chan string, 100) // Totally wrong, but whatever.
	go Publisher()

	go http.ListenAndServe(*statbindaddr, http.DefaultServeMux)
	fmt.Printf("2")

	log.SetLevel(log.InfoLevel)

	s := gobgp.NewBgpServer()
	go s.Serve()

	fmt.Printf("3")

	if *enablegrpc {
		// start grpc api server. this is not mandatory
		// but you will be able to use `gobgp` cmd with this.
		g := api.NewGrpcServer(s, ":50051")
		go g.Serve()
	}

	global := &config.Global{
		Config: config.GlobalConfig{
			As:       uint32(*selfasn),
			RouterId: *routerid,
			Port:     int32(*bgpport),
		},
	}

	if err := s.Start(global); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("4")

	peers := loadPeerConf()

	for _, peer := range peers {
		n := &config.Neighbor{
			Config: config.NeighborConfig{
				NeighborAddress: peer.Peeraddr,
				PeerAs:          uint32(peer.Peeras),
			},
		}

		n.EbgpMultihop.State.Enabled = true
		n.EbgpMultihop.State.MultihopTtl = 50
		n.Transport.State.LocalAddress = peer.Localaddr
		n.ApplyPolicy.State.DefaultExportPolicy = config.DEFAULT_POLICY_TYPE_REJECT_ROUTE

		n.EbgpMultihop.Config = config.EbgpMultihopConfig{
			Enabled:     true,
			MultihopTtl: 50,
		}

		n.Transport.Config = config.TransportConfig{
			LocalAddress: peer.Localaddr,
		}

		n.ApplyPolicy.Config = config.ApplyPolicyConfig{
			DefaultExportPolicy: config.DEFAULT_POLICY_TYPE_REJECT_ROUTE,
		}

		if err := s.AddNeighbor(n); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("5")
	}

	w := s.Watch(gobgp.WatchPostUpdate(true))
	for {
		select {
		case ev := <-w.Event():
			b, _ := json.Marshal(ev)
			PublishChan <- string(b)
			switch msg := ev.(type) {
			case *gobgp.WatchEventBestPath:
				for _, path := range msg.PathList {
					// do something useful
					fmt.Println(path)
				}
			}
		}
	}

	// // monitor new routes
	// req = gobgp.NewGrpcRequest(gobgp.REQ_MONITOR_RIB, "", bgp.RF_IPv4_UC, &api.Table{
	// 	Type: api.Resource_GLOBAL,
	// })
	// s.GrpcReqCh <- req

	// for res := range req.ResponseCh {
	// 	p, _ := cmd.ApiStruct2Path(res.Data.(*api.Destination).Paths[0])

	// 	// cmd.ShowRoute(p, false, false, false, true, false)
	// 	// api.Destination.Prefix
	// 	b, _ := json.Marshal(p)
	// 	PublishChan <- string(b)
	// 	for _, v := range p {
	// 		if v.IsWithdraw {
	// 			bgpWithdraws.Add(1)
	// 		}
	// 		bgpUpdates.Add(1)
	// 	}

	// }
}

func Publisher() {
	client := redis.NewClient(&redis.Options{
		Addr:     *redisaddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for msg := range PublishChan {
		if client.Publish(*redistopic, msg).Err() != nil {
			client.Close()
			client = redis.NewClient(&redis.Options{
				Addr:     *redisaddr,
				Password: "", // no password set
				DB:       0,  // use default DB
			})
		}
	}
}

func loadPeerConf() []BGPPeer {
	peers := make([]BGPPeer, 0)
	b, err := ioutil.ReadFile(*configfile)
	if err != nil {
		log.Fatalf("Unable to read bgp peer conf file: %s", err.Error())
	}

	err = json.Unmarshal(b, &peers)
	if err != nil {
		log.Fatalf("Unable to parse bgp peer conf file: %s", err.Error())
	}

	for n, peer := range peers {
		if peer.Localaddr == "" ||
			peer.Peeraddr == "" ||
			peer.Peeras == 0 {
			log.Fatalf("Peer %d has incorrect info, peer.Localaddr == '%s' , peer.Peeraddr == '%s', peer.Peeras == %d",
				n,
				peer.Localaddr,
				peer.Peeraddr,
				peer.Peeras)
		}
	}

	return peers

}
