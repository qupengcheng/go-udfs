package dht

import (
	"context"
	"fmt"
	"strings"

	cid "gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	notif "gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing/notifications"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	kb "gx/ipfs/QmesQqwonP618R7cJZoFfA4ioYhhMKnDmtUxcAvvxEEGnw/go-libp2p-kbucket"

	pb "github.com/ipfs/go-ipfs/udfs/go-libp2p-kad-dht/pb"
)

func tryFormatLoggableKey(k string) (string, error) {
	if len(k) == 0 {
		return "", fmt.Errorf("loggableKey is empty")
	}
	var proto, cstr string
	if k[0] == '/' {
		// it's a path (probably)
		protoEnd := strings.IndexByte(k[1:], '/')
		if protoEnd < 0 {
			return k, fmt.Errorf("loggableKey starts with '/' but is not a path: %x", k)
		}
		proto = k[1 : protoEnd+1]
		cstr = k[protoEnd+2:]
	} else {
		proto = "provider"
		cstr = k
	}

	c, err := cid.Cast([]byte(cstr))
	if err != nil {
		return "", fmt.Errorf("loggableKey could not cast key to a CID: %x %v", k, err)
	}
	return fmt.Sprintf("/%s/%s", proto, c.String()), nil
}

func loggableKey(k string) logging.LoggableMap {
	newKey, err := tryFormatLoggableKey(k)
	if err != nil {
		log.Debug(err)
	} else {
		k = newKey
	}

	return logging.LoggableMap{
		"key": k,
	}
}

// Kademlia 'node lookup' operation. Returns a channel of the K closest peers
// to the given key
func (dht *IpfsDHT) GetClosestPeers(ctx context.Context, key string) (<-chan peer.ID, error) {
	e := log.EventBegin(ctx, "getClosestPeers", loggableKey(key))
	tablepeers := dht.routingTable.NearestPeers(kb.ConvertKey(key), AlphaValue)
	if len(tablepeers) == 0 {
		return nil, kb.ErrLookupFailure
	}

	out := make(chan peer.ID, KValue)

	// since the query doesnt actually pass our context down
	// we have to hack this here. whyrusleeping isnt a huge fan of goprocess
	parent := ctx
	query := dht.newQuery(key, func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {
		// For DHT query command
		notif.PublishQueryEvent(parent, &notif.QueryEvent{
			Type: notif.SendingQuery,
			ID:   p,
		})

		pmes, err := dht.findPeerSingle(ctx, p, peer.ID(key))
		if err != nil {
			log.Debugf("error getting closer peers: %s", err)
			return nil, err
		}
		peers := pb.PBPeersToPeerInfos(pmes.GetCloserPeers())

		// For DHT query command
		notif.PublishQueryEvent(parent, &notif.QueryEvent{
			Type:      notif.PeerResponse,
			ID:        p,
			Responses: peers,
		})

		return &dhtQueryResult{closerPeers: peers}, nil
	})

	go func() {
		defer close(out)
		defer e.Done()
		// run it!
		res, err := query.Run(ctx, tablepeers)
		if err != nil {
			log.Debugf("closestPeers query run error: %s", err)
		}

		if res != nil && res.queriedSet != nil {
			sorted := kb.SortClosestPeers(res.queriedSet.Peers(), kb.ConvertKey(key))
			if len(sorted) > KValue {
				sorted = sorted[:KValue]
			}

			for _, p := range sorted {
				out <- p
			}
		}
	}()

	return out, nil
}

// Kademlia 'node lookup' operation. Returns a channel of the K closest peers
// to the given key
func (dht *IpfsDHT) GetClosestMasterPeers(ctx context.Context, key string) (<-chan peer.ID, error) {
	e := log.EventBegin(ctx, "getClosestMasterPeers", loggableKey(key))
	tablepeers := dht.routingTable.NearestMasterPeers(dht.peerstore, kb.ConvertKey(key), AlphaValue)
	if len(tablepeers) == 0 {
		return nil, kb.ErrLookupFailure
	}

	out := make(chan peer.ID, KValue)

	// since the query doesnt actually pass our context down
	// we have to hack this here. whyrusleeping isnt a huge fan of goprocess
	parent := ctx
	query := dht.newQuery(key, func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {
		// For DHT query command
		notif.PublishQueryEvent(parent, &notif.QueryEvent{
			Type: notif.SendingQuery,
			ID:   p,
		})

		pmes, err := dht.findMasterPeerSingle(ctx, p, peer.ID(key))
		if err != nil {
			log.Debugf("error getting closer peers: %s", err)
			return nil, err
		}
		peers := pb.PBPeersToPeerInfos(pmes.GetCloserPeers())

		// For DHT query command
		notif.PublishQueryEvent(parent, &notif.QueryEvent{
			Type:      notif.PeerResponse,
			ID:        p,
			Responses: peers,
		})

		return &dhtQueryResult{closerPeers: peers}, nil
	})

	go func() {
		defer close(out)
		defer e.Done()
		// run it!
		res, err := query.Run(ctx, tablepeers)
		if err != nil {
			log.Debugf("closestPeers query run error: %s", err)
		}

		if res != nil && res.queriedSet != nil {
			sorted := kb.SortClosestPeers(res.queriedSet.Peers(), kb.ConvertKey(key))
			if len(sorted) > KValue {
				sorted = sorted[:KValue]
			}

			for _, p := range sorted {
				out <- p
			}
		}
	}()

	return out, nil
}