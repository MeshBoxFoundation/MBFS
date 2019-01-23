package testutil

import (
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	ci "mbfs/go-mbfs/gx/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "mbfs/go-mbfs/gx/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
	mh "mbfs/go-mbfs/gx/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"
)

var generatedPairs int64 = 0

func RandPeerID() (peer.ID, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	h, _ := mh.Sum(buf, mh.SHA2_256, -1)
	return peer.ID(h), nil
}

func RandTestKeyPair(bits int) (ci.PrivKey, ci.PubKey, error) {
	seed := time.Now().UnixNano()

	// workaround for low time resolution
	seed += atomic.AddInt64(&generatedPairs, 1) << 32

	r := rand.New(rand.NewSource(seed))
	return ci.GenerateKeyPairWithReader(ci.RSA, bits, r)
}

func SeededTestKeyPair(seed int64) (ci.PrivKey, ci.PubKey, error) {
	r := rand.New(rand.NewSource(seed))
	return ci.GenerateKeyPairWithReader(ci.RSA, 512, r)
}
