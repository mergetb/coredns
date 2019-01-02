package nex

import (
	"context"
	"encoding/json"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
)

// all obejcts that go to and fro from the database implement this interface
type Object interface {
	Key() string
	GetVersion() int64
	SetVersion(int64)
	Value() interface{}
}

func Read(obj Object) error {

	return ReadObjects([]Object{obj})

}

func ReadObjects(objs []Object) error {

	var ops []etcd.Op
	omap := make(map[string]Object)

	for _, o := range objs {
		if !IndexExists(o) {
			continue
		}
		omap[o.Key()] = o
		ops = append(ops, etcd.OpGet(o.Key()))
	}

	return withEtcd(func(c *etcd.Client) error {

		kvc := etcd.NewKV(c)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			return TxnFailed("")
		}

		for _, r := range resp.Responses {
			rr := r.GetResponseRange()
			if rr == nil {
				continue
			}

			for _, kv := range rr.Kvs {
				o := omap[string(kv.Key)]
				err := json.Unmarshal(kv.Value, o.Value())
				if err != nil {
					return err
				}
				o.SetVersion(kv.Version)
			}
		}

		return nil

	})

}

func Write(obj Object) error {

	return WriteObjects([]Object{obj})

}

func WriteObjects(objs []Object) error {

	var ops []etcd.Op
	var ifs []etcd.Cmp

	for _, obj := range objs {

		if !IndexExists(obj) {
			continue
		}

		var value string
		switch t := obj.Value().(type) {
		case string:
			value = t
		default:
			buf, err := json.MarshalIndent(obj.Value(), "", "  ")
			if err != nil {
				return err
			}
			value = string(buf)
		}

		ops = append(ops, etcd.OpPut(obj.Key(), value))
		ifs = append(ifs,
			etcd.Compare(etcd.Version(obj.Key()), "=", obj.GetVersion()))

	}

	return withEtcd(func(c *etcd.Client) error {

		kvc := etcd.NewKV(c)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := kvc.Txn(ctx).If(ifs...).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			return TxnFailed("state has changed since read")
		}

		return nil

	})

}

func DeleteObjects(objs []Object) error {

	var ops []etcd.Op

	for _, obj := range objs {

		if !IndexExists(obj) {
			continue
		}

		ops = append(ops, etcd.OpDelete(obj.Key()))

	}

	return withEtcd(func(c *etcd.Client) error {

		kvc := etcd.NewKV(c)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			return TxnFailed("delete objects failed")
		}

		return nil

	})

}

type ObjectTx struct {
	Put    []Object
	Delete []Object
}

func RunObjectTx(otx ObjectTx) error {

	var ops []etcd.Op
	for _, x := range otx.Put {
		if !IndexExists(x) {
			continue
		}
		value, err := json.MarshalIndent(x.Value(), "", "  ")
		if err != nil {
			return nil
		}
		ops = append(ops, etcd.OpPut(x.Key(), string(value)))
	}
	for _, x := range otx.Delete {
		if !IndexExists(x) {
			continue
		}
		ops = append(ops, etcd.OpDelete(x.Key()))
	}

	return withEtcd(func(c *etcd.Client) error {

		kvc := etcd.NewKV(c)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			return TxnFailed("run object txn failed")
		}

		return nil

	})

}
