package test

import (
	"context"
	"fmt"
	"testing"

	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/require"
	ipfslog "github.com/thusharprakash/go-ipfs-log"
	idp "github.com/thusharprakash/go-ipfs-log/identityprovider"
	"github.com/thusharprakash/go-ipfs-log/keystore"
	mocknet "github.com/thusharprakash/go-libp2p/p2p/net/mock"
)

func BenchmarkAdd(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := mocknet.New()
	defer m.Close()

	ipfs, closeNode := NewMemoryServices(ctx, b, m)
	defer closeNode()

	datastore := dssync.MutexWrap(NewIdentityDataStore(b))
	ks, err := keystore.NewKeystore(datastore)
	require.NoError(b, err)

	identity, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
		Keystore: ks,
		ID:       "userA",
		Type:     "orbitdb",
	})

	log, err := ipfslog.NewLog(ipfs, identity, &ipfslog.LogOptions{ID: "A"})
	require.NoError(b, err)

	b.ResetTimer()
	// Start the main loop
	for n := 0; n < b.N; n++ {
		_, err = log.Append(ctx, []byte(fmt.Sprintf("%d", n)), nil)
		require.NoError(b, err)
	}
}
