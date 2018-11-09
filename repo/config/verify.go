package config

type VerifyInfo struct {
	ServerAddress string
	ServerPubkey  string

	Txid       string
	Voutid     int32
	Licversion int32
	Secret     string

	License string
	Period  int64
}
