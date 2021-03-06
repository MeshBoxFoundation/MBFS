package libp2p

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	crypto "mbfs/go-mbfs/gx/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	pstore "mbfs/go-mbfs/gx/QmUymf8fJtideyv3z727BcZUifGBjMZMpCJqu3Gxk5aRUk/go-libp2p-peerstore"
	host "mbfs/go-mbfs/gx/QmVrjR2KMe57y4YyfHdYa3yKD278gN8W7CTiqSuYmxjA7F/go-libp2p-host"
	"mbfs/go-mbfs/gx/QmetLkVrNTi6QGfv577kpKPVBXJFshKZrP87C7xHJp7SCX/go-tcp-transport"
)

func TestNewHost(t *testing.T) {
	h, err := makeRandomHost(t, 9000)
	if err != nil {
		t.Fatal(err)
	}
	h.Close()
}

func TestBadTransportConstructor(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx, Transport(func() {}))
	if err == nil {
		h.Close()
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "libp2p_test.go") {
		t.Error("expected error to contain debugging info")
	}
}

func TestNoListenAddrs(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx, NoListenAddrs)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	if len(h.Addrs()) != 0 {
		t.Fatal("expected no addresses")
	}
}

func TestNoTransports(t *testing.T) {
	ctx := context.Background()
	a, err := New(ctx, NoTransports)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	b, err := New(ctx, ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	err = a.Connect(ctx, pstore.PeerInfo{
		ID:    b.ID(),
		Addrs: b.Addrs(),
	})
	if err == nil {
		t.Error("dial should have failed as no transports have been configured")
	}
}

func TestInsecure(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx, NoSecurity)
	if err != nil {
		t.Fatal(err)
	}
	h.Close()
}

func TestDefaultListenAddrs(t *testing.T) {
	ctx := context.Background()

	re := regexp.MustCompile("/(ip)[4|6]/((0.0.0.0)|(::))/tcp/")
	re2 := regexp.MustCompile("/p2p-circuit")

	// Test 1: Setting the correct listen addresses if userDefined.Transport == nil && userDefined.ListenAddrs == nil
	h, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range h.Network().ListenAddresses() {
		if re.FindStringSubmatchIndex(addr.String()) == nil &&
			re2.FindStringSubmatchIndex(addr.String()) == nil {
			t.Error("expected ip4 or ip6 or relay interface")
		}
	}

	h.Close()

	// Test 2: Listen addr only include relay if user defined transport is passed.
	h, err = New(
		ctx,
		Transport(tcp.NewTCPTransport),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(h.Network().ListenAddresses()) != 1 {
		t.Error("expected one listen addr with user defined transport")
	}
	if re2.FindStringSubmatchIndex(h.Network().ListenAddresses()[0].String()) == nil {
		t.Error("expected relay address")
	}
	h.Close()
}

func makeRandomHost(t *testing.T, port int) (host.Host, error) {
	ctx := context.Background()
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		t.Fatal(err)
	}

	opts := []Option{
		ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		Identity(priv),
		DefaultTransports,
		DefaultMuxers,
		DefaultSecurity,
		NATPortMap(),
	}

	return New(ctx, opts...)
}
