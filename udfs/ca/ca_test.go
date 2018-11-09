package ca

import (
	"fmt"
	"testing"
)

var (
	txid   = "ebfad1c5e579b42837625b421d58420f4d41cb0b2c488e74ad0bd9d5291948eb"
	voutid = int32(0)

	period = int64(1541784038)

	licversion = int32(2)

	privStr   = "5JCRtN29x3QRxY5uyijQGbmyU7e1pN8EXWcifTsfjqKFnYUefyZ"
	pubkeyStr = "04626a7f8decd5a7fd8a862101a2b9e5f78ac91e2132884c293bea3f983905eced14e708dd2819d06e19fb237324b9b3bfbac4e8b8d61e1ac313b4d41dbdb0bf5f"
	hashStr   = "353b4297f772e6b2ee191aec6b06f526464a81a058fc70f84898504ab639a00b"

	signStr = "INJPgVEfoKnWMr7qGucRCTDPfi6Jo7OF9EQ0lQn1A/ZbDjJkLsLNh3k9u58zk0xzo6bU//qmiU3Gu2mlJ1jEhZ8="

	ucenterPubkey = "03a00f7bf6cf623a7b5aba1b8e5086c05faa9a59e8a0f70a46bea2a2590fd00b95"

	testServerAddress = ""
)

func Test_MakeBTCHash(t *testing.T) {
	hash := MakeNodeInfoHash(txid, voutid, pubkeyStr, period, licversion)
	if hash != hashStr {
		t.FailNow()
	}
}

func Test_RequestSignature(t *testing.T) {
	lbi, e := RequestLicense(testServerAddress, txid, voutid)
	if e != nil {
		t.Fatal(e)
	}

	if lbi.LicPeriod != period {
		t.FailNow()
	}

	if lbi.License != signStr {
		t.FailNow()
	}

}

func Test_VerifySignature(t *testing.T) {
	ok, err := VerifySignature(hashStr, signStr, ucenterPubkey)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.FailNow()
	}
}

func Test_MakePrivateAddr(t *testing.T) {
	fmt.Println(MakePrivateAddr())
}

func Test_PublicKeyFromPrivateAddr(t *testing.T) {
	pub, err := PublicKeyFromPrivateAddr(privStr)
	if err != nil {
		t.Fatal(err)
	}

	if pub != pubkeyStr {
		t.FailNow()
	}
}

func Test_Sign(t *testing.T) {
	hash := MakeRandomHash()
	b64, err := Sign(hash, privStr)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := VerifySignature(hash, b64, pubkeyStr)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.FailNow()
	}
}
