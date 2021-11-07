package keyring

import (
	kw "github.com/zalando/go-keyring/kwallet"
	ss "github.com/zalando/go-keyring/secret_service"
)

func init() {
	// default to secret service and fall back to kwallet — most systems will only
	// have one of the two available anyways
	secretService, err := ss.NewSecretService()
	if err == nil {
		provider = secretService
		return
	}
	kwallet, err := kw.NewKWallet()
	if err == nil {
		provider = kwallet
		return
	}
}
