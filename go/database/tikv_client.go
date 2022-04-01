package database

import (
	"bytes"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/terror"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"

	goctx "golang.org/x/net/context"
)

const (
	PASSWORD_KEY = "password_sm3:"
	CODE_KEY     = "code:"
)

type KV struct {
	K, V []byte
}

var (
	store kv.Storage
)

func TiKVClientInit(pdAddr string) {
	driver := tikv.Driver{}
	var err error
	store, err = driver.Open(pdAddr)
	terror.MustNil(err)
}

func TiKVClientGet(k []byte) (KV, error) {
	tx, err := store.Begin()
	if err != nil {
		return KV{}, errors.Trace(err)
	}
	v, err := tx.Get(goctx.Background(), k)
	if err != nil {
		return KV{}, errors.Trace(err)
	}
	return KV{K: k, V: v}, nil
}

func TiKVClientDeletes(keys ...[]byte) error {
	tx, err := store.Begin()
	if err != nil {
		return errors.Trace(err)
	}
	for _, key := range keys {
		err := tx.Delete(key)
		if err != nil {
			return errors.Trace(err)
		}
	}
	err = tx.Commit(goctx.Background())
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func TiKVClientPuts(args ...[]byte) error {
	tx, err := store.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	for i := 0; i < len(args); i += 2 {
		key, val := args[i], args[i+1]
		err := tx.Set(key, val)
		if err != nil {
			return errors.Trace(err)
		}
	}
	err = tx.Commit(goctx.Background())
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func TiKVClientScan(keyPrefix []byte, limit int) ([]KV, error) {
	tx, err := store.Begin()
	if err != nil {
		return nil, errors.Trace(err)
	}
	it, err := tx.Iter(kv.Key(keyPrefix), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer it.Close()
	var ret []KV
	for it.Valid() && limit > 0 {
		ret = append(ret, KV{K: it.Key()[:], V: it.Value()[:]})
		limit--
		it.Next()
	}
	return ret, nil
}

func TiKVClientUpdate(k []byte, v []byte) ([]byte, error) {
	tx, err := store.Begin()
	if err != nil {
		return nil, errors.Trace(err)
	}

	oldValue, err := tx.Get(goctx.Background(), k)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !bytes.Equal(oldValue, v) {
		err = tx.Set(k, v)
		if err != nil {
			return nil, errors.Trace(err)
		}

		err = tx.Commit(goctx.Background())
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return oldValue, nil
}
