package metrics

import (
	protocol "mbfs/go-mbfs/gx/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "mbfs/go-mbfs/gx/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
)

type Reporter interface {
	LogSentMessage(int64)
	LogRecvMessage(int64)
	LogSentMessageStream(int64, protocol.ID, peer.ID)
	LogRecvMessageStream(int64, protocol.ID, peer.ID)
	GetBandwidthForPeer(peer.ID) Stats
	GetBandwidthForProtocol(protocol.ID) Stats
	GetBandwidthTotals() Stats
}
