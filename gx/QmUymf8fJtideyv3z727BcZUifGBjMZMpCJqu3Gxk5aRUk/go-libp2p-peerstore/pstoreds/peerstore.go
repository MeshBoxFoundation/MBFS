package pstoreds

import (
	"context"
	"time"

	base32 "mbfs/go-mbfs/gx/QmfVj3x4D6Jkq9SEoi5n2NmoUomLwoeiwnYz2KQa15wRw6/base32"

	ds "mbfs/go-mbfs/gx/QmaRb5yNXKonhbkpNxNawoydk4N6es6b4fPj19sjEKsh5D/go-datastore"
	query "mbfs/go-mbfs/gx/QmaRb5yNXKonhbkpNxNawoydk4N6es6b4fPj19sjEKsh5D/go-datastore/query"

	pstore "mbfs/go-mbfs/gx/QmUymf8fJtideyv3z727BcZUifGBjMZMpCJqu3Gxk5aRUk/go-libp2p-peerstore"
	peer "mbfs/go-mbfs/gx/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
)

// Configuration object for the peerstore.
type Options struct {
	// The size of the in-memory cache. A value of 0 or lower disables the cache.
	CacheSize uint

	// Sweep interval to expire entries, only used when TTL is *not* natively managed
	// by the underlying datastore.
	TTLInterval time.Duration

	// Number of times to retry transactional writes.
	WriteRetries uint
}

// DefaultOpts returns the default options for a persistent peerstore:
// * Cache size: 1024
// * TTL sweep interval: 1 second
// * WriteRetries: 5
func DefaultOpts() Options {
	return Options{
		CacheSize:    1024,
		TTLInterval:  time.Second,
		WriteRetries: 5,
	}
}

// NewPeerstore creates a peerstore backed by the provided persistent datastore.
func NewPeerstore(ctx context.Context, store ds.TxnDatastore, opts Options) (pstore.Peerstore, error) {
	addrBook, err := NewAddrBook(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	keyBook, err := NewKeyBook(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	peerMetadata, err := NewPeerMetadata(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	ps := pstore.NewPeerstore(keyBook, addrBook, peerMetadata)
	return ps, nil
}

// uniquePeerIds extracts and returns unique peer IDs from database keys.
func uniquePeerIds(ds ds.TxnDatastore, prefix ds.Key, extractor func(result query.Result) string) (peer.IDSlice, error) {
	var (
		q       = query.Query{Prefix: prefix.String(), KeysOnly: true}
		results query.Results
		err     error
	)

	txn, err := ds.NewTransaction(true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	if results, err = txn.Query(q); err != nil {
		log.Error(err)
		return nil, err
	}

	defer results.Close()

	idset := make(map[string]struct{})
	for result := range results.Next() {
		k := extractor(result)
		idset[k] = struct{}{}
	}

	if len(idset) == 0 {
		return peer.IDSlice{}, nil
	}

	ids := make(peer.IDSlice, len(idset))
	i := 0
	for id := range idset {
		pid, _ := base32.RawStdEncoding.DecodeString(id)
		ids[i], _ = peer.IDFromBytes(pid)
		i++
	}
	return ids, nil
}