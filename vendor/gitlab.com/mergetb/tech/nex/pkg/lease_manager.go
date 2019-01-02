package nex

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	quantum = 1 * time.Minute
)

func RunLeaseManager() {

	for ; ; time.Sleep(quantum) {

		members, err := FetchIp4IndexMembers()
		if err != nil {
			log.WithError(err).Errorf("lease-manager: fetch ip4 index failed")
			continue
		}

		err = RecycleExpiredLeases(members)
		if err != nil {
			log.WithError(err).Errorf("lease-manager: recycle failed")
			continue
		}
	}

}

func RecycleExpiredLeases(members []*Member) error {

	var updates, trash []Object
	for _, m := range members {

		if m.Ip4 == nil {
			continue
		}
		if m.Ip4.Expires == nil {
			continue
		}

		expires := time.Unix(m.Ip4.Expires.Seconds, int64(m.Ip4.Expires.Nanos))
		if time.Now().After(expires) {

			log.WithFields(log.Fields{
				"member": m.Mac,
				"addr":   m.Ip4.Address,
			}).Info("recycling expired address")

			trash = append(trash, NewIp4Index(m))
			u := m.Clone()
			u.Ip4 = nil
			updates = append(updates, NewMacIndex(u))

		}

	}

	return RunObjectTx(ObjectTx{Put: updates, Delete: trash})

}
