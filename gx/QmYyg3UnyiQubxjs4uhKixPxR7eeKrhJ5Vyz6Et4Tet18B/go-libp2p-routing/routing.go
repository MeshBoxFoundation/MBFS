// package routing defines the interface for a routing system used by ipfs.
package routing

import (
	"context"
	"errors"

	ropts "mbfs/go-mbfs/gx/QmYyg3UnyiQubxjs4uhKixPxR7eeKrhJ5Vyz6Et4Tet18B/go-libp2p-routing/options"

	ci "mbfs/go-mbfs/gx/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	cid "mbfs/go-mbfs/gx/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	pstore "mbfs/go-mbfs/gx/QmUymf8fJtideyv3z727BcZUifGBjMZMpCJqu3Gxk5aRUk/go-libp2p-peerstore"
	peer "mbfs/go-mbfs/gx/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
)

// ErrNotFound is returned when the router fails to find the requested record.
var ErrNotFound = errors.New("routing: not found")

// ErrNotSupported is returned when the router doesn't support the given record
// type/operation.
var ErrNotSupported = errors.New("routing: operation or key not supported")

// ContentRouting is a value provider layer of indirection. It is used to find
// information about who has what content.
type ContentRouting interface {
	// Provide adds the given cid to the content routing system. If 'true' is
	// passed, it also announces it, otherwise it is just kept in the local
	// accounting of which objects are being provided.
	Provide(context.Context, cid.Cid, bool) error

	// Search for peers who are able to provide a given key
	FindProvidersAsync(context.Context, cid.Cid, int) <-chan pstore.PeerInfo
}

// PeerRouting is a way to find information about certain peers.
// This can be implemented by a simple lookup table, a tracking server,
// or even a DHT.
type PeerRouting interface {
	// Find specific Peer
	// FindPeer searches for a peer with given ID, returns a pstore.PeerInfo
	// with relevant addresses.
	FindPeer(context.Context, peer.ID) (pstore.PeerInfo, error)
}

// ValueStore is a basic Put/Get interface.
type ValueStore interface {

	// PutValue adds value corresponding to given Key.
	PutValue(context.Context, string, []byte, ...ropts.Option) error

	// GetValue searches for the value corresponding to given Key.
	GetValue(context.Context, string, ...ropts.Option) ([]byte, error)

	// SearchValue searches for better and better values from this value
	// store corresponding to the given Key. By default implementations must
	// stop the search after a good value is found. A 'good' value is a value
	// that would be returned from GetValue.
	//
	// Useful when you want a result *now* but still want to hear about
	// better/newer results.
	//
	// Implementations of this methods won't return ErrNotFound. When a value
	// couldn't be found, the channel will get closed without passing any results
	SearchValue(context.Context, string, ...ropts.Option) (<-chan []byte, error)
}

// IpfsRouting is the combination of different routing types that ipfs
// uses. It can be satisfied by a single item (such as a DHT) or multiple
// different pieces that are more optimized to each task.
type IpfsRouting interface {
	ContentRouting
	PeerRouting
	ValueStore

	// Bootstrap allows callers to hint to the routing system to get into a
	// Boostrapped state
	Bootstrap(context.Context) error

	// TODO expose io.Closer or plain-old Close error
}

// PubKeyFetcher is an interfaces that should be implemented by value stores
// that can optimize retrieval of public keys.
//
// TODO(steb): Consider removing, see #22.
type PubKeyFetcher interface {
	// GetPublicKey returns the public key for the given peer.
	GetPublicKey(context.Context, peer.ID) (ci.PubKey, error)
}

// KeyForPublicKey returns the key used to retrieve public keys
// from a value store.
func KeyForPublicKey(id peer.ID) string {
	return "/pk/" + string(id)
}

// GetPublicKey retrieves the public key associated with the given peer ID from
// the value store.
//
// If the ValueStore is also a PubKeyFetcher, this method will call GetPublicKey
// (which may be better optimized) instead of GetValue.
func GetPublicKey(r ValueStore, ctx context.Context, p peer.ID) (ci.PubKey, error) {
	switch k, err := p.ExtractPublicKey(); err {
	case peer.ErrNoPublicKey:
		// check the datastore
	case nil:
		return k, nil
	default:
		return nil, err
	}

	if dht, ok := r.(PubKeyFetcher); ok {
		// If we have a DHT as our routing system, use optimized fetcher
		return dht.GetPublicKey(ctx, p)
	}
	key := KeyForPublicKey(p)
	pkval, err := r.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}

	// get PublicKey from node.Data
	return ci.UnmarshalPublicKey(pkval)
}
