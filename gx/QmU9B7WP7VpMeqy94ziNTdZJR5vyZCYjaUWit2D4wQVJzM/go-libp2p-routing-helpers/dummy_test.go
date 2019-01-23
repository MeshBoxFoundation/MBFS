package routinghelpers

import (
	"context"
	"errors"
	"strings"
	"sync"

	cid "mbfs/go-mbfs/gx/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	pstore "mbfs/go-mbfs/gx/QmUymf8fJtideyv3z727BcZUifGBjMZMpCJqu3Gxk5aRUk/go-libp2p-peerstore"
	routing "mbfs/go-mbfs/gx/QmYyg3UnyiQubxjs4uhKixPxR7eeKrhJ5Vyz6Et4Tet18B/go-libp2p-routing"
	ropts "mbfs/go-mbfs/gx/QmYyg3UnyiQubxjs4uhKixPxR7eeKrhJ5Vyz6Et4Tet18B/go-libp2p-routing/options"
	peer "mbfs/go-mbfs/gx/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
)

type dummyValueStore sync.Map

func (d *dummyValueStore) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	if strings.HasPrefix(key, "/notsupported/") {
		return routing.ErrNotSupported
	}
	if strings.HasPrefix(key, "/error/") {
		return errors.New(key[len("/error/"):])
	}
	if strings.HasPrefix(key, "/stall/") {
		<-ctx.Done()
		return ctx.Err()
	}
	(*sync.Map)(d).Store(key, value)
	return nil
}

func (d *dummyValueStore) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	if strings.HasPrefix(key, "/error/") {
		return nil, errors.New(key[len("/error/"):])
	}
	if strings.HasPrefix(key, "/stall/") {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if v, ok := (*sync.Map)(d).Load(key); ok {
		return v.([]byte), nil
	}
	return nil, routing.ErrNotFound
}

func (d *dummyValueStore) SearchValue(ctx context.Context, key string, opts ...ropts.Option) (<-chan []byte, error) {
	out := make(chan []byte)
	if strings.HasPrefix(key, "/error/") {
		return nil, errors.New(key[len("/error/"):])
	}

	go func() {
		defer close(out)
		v, err := d.GetValue(ctx, key, opts...)
		if err == nil {
			select {
			case out <- v:
			case <-ctx.Done():
			}
		}
	}()
	return out, nil
}

type dummyProvider map[string][]peer.ID

func (d dummyProvider) FindProvidersAsync(ctx context.Context, c cid.Cid, count int) <-chan pstore.PeerInfo {
	peers := d[c.KeyString()]
	if len(peers) > count {
		peers = peers[:count]
	}
	out := make(chan pstore.PeerInfo)
	go func() {
		defer close(out)
		for _, p := range peers {
			if p == "stall" {
				<-ctx.Done()
				return
			}
			select {
			case out <- pstore.PeerInfo{ID: p}:
			case <-ctx.Done():
			}
		}
	}()
	return out
}

func (d dummyProvider) Provide(ctx context.Context, c cid.Cid, local bool) error {
	return routing.ErrNotSupported
}

type cbProvider func(c cid.Cid, local bool) error

func (d cbProvider) Provide(ctx context.Context, c cid.Cid, local bool) error {
	return d(c, local)
}

func (d cbProvider) FindProvidersAsync(ctx context.Context, c cid.Cid, count int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	close(ch)
	return ch
}

type dummyPeerRouter map[peer.ID]struct{}

func (d dummyPeerRouter) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	if _, ok := d[p]; ok {
		return pstore.PeerInfo{ID: p}, nil
	}
	return pstore.PeerInfo{}, routing.ErrNotFound
}
