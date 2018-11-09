package ca

import (
	"fmt"
	"testing"
)

var (
	txid   = "ebfad1c5e579b42837625b421d58420f4d41cb0b2c488e74ad0bd9d5291948eb"
	voutid = int32(0)

	period = int64(1541751609)

	licversion = int32(2)

	privStr   = "5JCRtN29x3QRxY5uyijQGbmyU7e1pN8EXWcifTsfjqKFnYUefyZ"
	pubkeyStr = "04626a7f8decd5a7fd8a862101a2b9e5f78ac91e2132884c293bea3f983905eced14e708dd2819d06e19fb237324b9b3bfbac4e8b8d61e1ac313b4d41dbdb0bf5f"
	hashStr   = "fbcfd1b7d8b37e0fcd68c0116eac50d8a295aba0c94ab9ca7a029c63f6db0f89"

	signStr = "IITnhcC1LOmpHr0LEq4SPgJSWSJEDUPgUJSOIYmrp0JSSCMGS9yxGH13hjkb4gVYcFoxVR0blBunVOhgX3ub+VA="
)

func Test_MakeBTCHash(t *testing.T) {
	txid = "f242eb1d0c1b908d948ddf0a145bb348e91297c871c2a5537a0e15e93fc57d37"
	voutid = 1
	pubkeyStr = "0447f4019f73dd7941283794fe1a54ef38c7a327a8b6f603fd736b64495fb77ddb9087ba72285592889db8e4c58732478b091e70148e30f77c71480f6b6788aa18"
	period = 1542207211
	licversion = 1
	hash := MakeNodeInfoHash(txid, voutid, pubkeyStr, period, licversion)
	fmt.Println(hash)
	if hash != hashStr {
		t.FailNow()
	}
}

func Test_RequestSignature(t *testing.T) {
	lbi, e := RequestLicense(txid, voutid)
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
	hashStr = "35f48c75a15e2dbdb38b68f967860477bc0746838ca677732a41f386bac692c3"
	signStr = "HzwzJX9Lxh9/8HCLLSuQddFr3jG1gX51b8H1ORB8NtsPbCP5KmZuTVUjtim4Y55NfcRTU3q8I9LmDlTl/0hapV8="
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
