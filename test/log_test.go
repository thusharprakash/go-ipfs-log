package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/require"
	ipfslog "github.com/thusharprakash/go-ipfs-log"
	"github.com/thusharprakash/go-ipfs-log/entry"
	"github.com/thusharprakash/go-ipfs-log/errmsg"
	idp "github.com/thusharprakash/go-ipfs-log/identityprovider"
	"github.com/thusharprakash/go-ipfs-log/iface"
	ks "github.com/thusharprakash/go-ipfs-log/keystore"
	mocknet "github.com/thusharprakash/go-libp2p/p2p/net/mock"
)

func TestLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := mocknet.New()
	defer m.Close()
	ipfs, closeNode := NewMemoryServices(ctx, t, m)
	defer closeNode()

	datastore := dssync.MutexWrap(NewIdentityDataStore(t))
	keystore, err := ks.NewKeystore(datastore)
	require.NoError(t, err)

	var identities []*idp.Identity

	for i := 0; i < 4; i++ {
		char := 'A' + i

		identity, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
			Keystore: keystore,
			ID:       fmt.Sprintf("user%c", char),
			Type:     "orbitdb",
		})
		require.NoError(t, err)

		identities = append(identities, identity)
	}

	t.Run("sets an id and a clock id", func(t *testing.T) {
		log1, err := ipfslog.NewLog(ipfs, identities[0], &ipfslog.LogOptions{ID: "A"})
		require.NoError(t, err)
		require.Equal(t, log1.ID, "A")
		require.Equal(t, log1.Clock.GetID(), identities[0].PublicKey)
	})

	t.Run("sets time.now as id string if id is not passed as an argument", func(t *testing.T) {
		before := time.Now().Unix() / 1000
		log1, err := ipfslog.NewLog(ipfs, identities[0], nil)
		require.NoError(t, err)
		after := time.Now().Unix() / 1000

		logid, err := strconv.ParseInt(log1.ID, 10, 64)
		require.NoError(t, err)

		require.GreaterOrEqual(t, logid, before)
		require.LessOrEqual(t, logid, after)
	})

	t.Run("sets items if given as params", func(t *testing.T) {
		id1, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
			Keystore: keystore,
			ID:       "userA",
			Type:     "orbitdb",
		})
		require.NoError(t, err)

		id2, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
			Keystore: keystore,
			ID:       "userB",
			Type:     "orbitdb",
		})
		require.NoError(t, err)

		id3, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
			Keystore: keystore,
			ID:       "userC",
			Type:     "orbitdb",
		})
		require.NoError(t, err)

		e1, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryA"), LogID: "A", Clock: entry.NewLamportClock(id1.PublicKey, 0)}, nil)
		require.NoError(t, err)

		e2, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryB"), LogID: "A", Clock: entry.NewLamportClock(id2.PublicKey, 1)}, nil)
		require.NoError(t, err)

		e3, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryC"), LogID: "A", Clock: entry.NewLamportClock(id3.PublicKey, 2)}, nil)
		require.NoError(t, err)

		log1, err := ipfslog.NewLog(ipfs, identities[0], &ipfslog.LogOptions{ID: "A", Entries: entry.NewOrderedMapFromEntries([]iface.IPFSLogEntry{e1, e2, e3})})
		require.NoError(t, err)

		values := log1.Values()

		require.Equal(t, values.Len(), 3)

		keys := values.Keys()
		require.Equal(t, string(values.UnsafeGet(keys[0]).GetPayload()), "entryA")
		require.Equal(t, string(values.UnsafeGet(keys[1]).GetPayload()), "entryB")
		require.Equal(t, string(values.UnsafeGet(keys[2]).GetPayload()), "entryC")
	})

	t.Run("sets heads if given as params", func(t *testing.T) {
		e1, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryA"), LogID: "A"}, nil)
		require.NoError(t, err)

		e2, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryB"), LogID: "A"}, nil)
		require.NoError(t, err)

		e3, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryC"), LogID: "A"}, nil)
		require.NoError(t, err)

		log1, err := ipfslog.NewLog(ipfs, identities[0], &ipfslog.LogOptions{ID: "B", Entries: entry.NewOrderedMapFromEntries([]iface.IPFSLogEntry{e1, e2, e3}), Heads: []iface.IPFSLogEntry{e3}})
		require.NoError(t, err)

		heads := log1.Heads()
		require.Equal(t, heads.Len(), 1)

		headsKeys := heads.Keys()
		require.Equal(t, heads.UnsafeGet(headsKeys[0]).GetHash().String(), e3.GetHash().String())
	})

	t.Run("finds heads if heads not given as params", func(t *testing.T) {
		e1, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryA"), LogID: "A"}, nil)
		require.NoError(t, err)

		e2, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryB"), LogID: "A"}, nil)
		require.NoError(t, err)

		e3, err := entry.CreateEntry(ctx, ipfs, identities[0], &entry.Entry{Payload: []byte("entryC"), LogID: "A"}, nil)
		require.NoError(t, err)

		log1, err := ipfslog.NewLog(ipfs, identities[0], &ipfslog.LogOptions{ID: "A", Entries: entry.NewOrderedMapFromEntries([]iface.IPFSLogEntry{e1, e2, e3})})
		require.NoError(t, err)

		heads := log1.Heads()
		require.Equal(t, heads.Len(), 3)

		headsKeys := heads.Keys()
		require.Equal(t, heads.UnsafeGet(headsKeys[2]).GetHash().String(), e1.GetHash().String())
		require.Equal(t, heads.UnsafeGet(headsKeys[1]).GetHash().String(), e2.GetHash().String())
		require.Equal(t, heads.UnsafeGet(headsKeys[0]).GetHash().String(), e3.GetHash().String())
	})

	t.Run("creates default public AccessController if not defined", func(t *testing.T) {
		log1, err := ipfslog.NewLog(ipfs, identities[0], nil)
		require.NoError(t, err)

		err = log1.AccessController.CanAppend(&entry.Entry{Payload: []byte("any")}, identities[0].Provider, nil)
		require.NoError(t, err)
	})

	t.Run("returns an error if ipfs is not net", func(t *testing.T) {
		log1, err := ipfslog.NewLog(nil, identities[0], nil)
		require.Nil(t, log1)
		require.Equal(t, err, errmsg.ErrIPFSNotDefined)
	})

	t.Run("returns an error if identity is not net", func(t *testing.T) {
		log1, err := ipfslog.NewLog(ipfs, nil, nil)
		require.Nil(t, log1)
		require.Equal(t, err, errmsg.ErrIdentityNotDefined)
	})

	t.Run("toString", func(t *testing.T) {
		expectedData := "five\n└─four\n  └─three\n    └─two\n      └─one"
		log1, err := ipfslog.NewLog(ipfs, identities[0], &ipfslog.LogOptions{ID: "A"})
		require.NoError(t, err)
		for _, val := range []string{"one", "two", "three", "four", "five"} {
			_, err := log1.Append(ctx, []byte(val), nil)
			require.NoError(t, err)
		}

		require.Equal(t, log1.ToString(nil), expectedData)
	})
}
