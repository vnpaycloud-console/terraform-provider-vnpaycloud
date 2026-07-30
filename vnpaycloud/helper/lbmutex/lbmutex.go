package lbmutex

import "terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"

func Key(lbID string) string {
	return "vnpaycloud_lb/" + lbID
}

func Lock(mkv *mutexkv.MutexKV, lbID string) func() {
	if lbID == "" {
		return func() {}
	}
	key := Key(lbID)
	mkv.Lock(key)
	return func() { mkv.Unlock(key) }
}
