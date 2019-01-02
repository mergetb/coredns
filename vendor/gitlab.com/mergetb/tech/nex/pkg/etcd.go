package nex

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	log "github.com/sirupsen/logrus"
)

func withEtcd(f func(*clientv3.Client) error) error {
	c, err := EtcdClient()
	if err != nil {
		return err
	}
	defer c.Close()

	return f(c)
}

func EtcdClient() (*clientv3.Client, error) {
	err := LoadConfig()
	if err != nil {
		return nil, err
	}
	c := Current.Etcd

	capool := x509.NewCertPool()
	capem, err := ioutil.ReadFile(c.CAcert)
	if err != nil {
		return nil, Errorf(fmt.Sprintf("error reading cacert '%s'", c.CAcert), err)
	}
	ok := capool.AppendCertsFromPEM(capem)
	if !ok {
		log.Error("ca invalid")
		return nil, fmt.Errorf("ca invalid")
	}

	cert, err := tls.LoadX509KeyPair(
		c.Cert,
		c.Key,
	)
	if err != nil {
		log.Errorf("error loading keys: %s", err)
		return nil, err
	}

	tlsc := &tls.Config{
		RootCAs:      capool,
		Certificates: []tls.Certificate{cert},
	}

	connstr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{connstr},
		DialTimeout: 2 * time.Second,
		TLS:         tlsc,
	})

	return cli, err
}

func fetchOneKV(format string, args ...interface{}) (*mvccpb.KeyValue, error) {
	c, err := EtcdClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	key := fmt.Sprintf(format, args...)
	resp, err := c.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0], nil
}

func fetchKvs(format string, args ...interface{}) ([]*mvccpb.KeyValue, error) {
	c, err := EtcdClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	key := fmt.Sprintf(format, args...)
	resp, err := c.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return resp.Kvs, nil
}
