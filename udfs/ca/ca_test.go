package ca

import (
	"fmt"
	"testing"
)

var (
	txid   = "f242eb1d0c1b908d948ddf0a145bb348e91297c871c2a5537a0e15e93fc57d37"
	voutid = int32(1)

	period     = int64(1542207211)
	licversion = int32(1)

	privStr   = "5KQPWmjrtWhvwRhxRiAsSwCdCYr4gGnv1MuhQuBBFa6wq2pu854"
	pubkeyStr = "0447f4019f73dd7941283794fe1a54ef38c7a327a8b6f603fd736b64495fb77ddb9087ba72285592889db8e4c58732478b091e70148e30f77c71480f6b6788aa18"
	hashStr   = "35f48c75a15e2dbdb38b68f967860477bc0746838ca677732a41f386bac692c3"

	signStr = "HzwzJX9Lxh9/8HCLLSuQddFr3jG1gX51b8H1ORB8NtsPbCP5KmZuTVUjtim4Y55NfcRTU3q8I9LmDlTl/0hapV8="

	ucenterPubkeyStr = "03e947099921ee170da47a7acf48143c624d33950af362fc39a734b1b3188ec1e3"
)

func Test_MakeBTCHash(t *testing.T) {
	hash := MakeNodeInfoHash(txid, voutid, pubkeyStr, period, licversion)
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
	ok, err := VerifySignature(hashStr, signStr, ucenterPubkeyStr)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.FailNow()
	}
}

//
//func Test_Raw(t *testing.T) {
//	b := bytes.NewBuffer(nil)
//	// strMessageMagic
//	strMessageMagic := "Ulord Signed Message:\n"
//	//WriteVlen(b, uint64(len(strMessageMagic)))
//	//b.Write([]byte(strMessageMagic))
//	writeString(b, strMessageMagic)
//
//	// _txid
//	//txid := "da29e46c4d805474cbed1e6771473c77b350858e990de26837fcbeb8a0271acf"
//	//WriteVlen(b, uint64(len(txid)))
//	//b.Write([]byte(txid))
//	writeString(b, txid)
//
//	// _voutid
//	//voutid := uint32(0)
//	binary.Write(b, binary.LittleEndian, uint32(voutid))
//
//	// _pubkey
//	//pubkeyStr := "041df129b64bf4091c6093b7c6253479b96541ce1fef817a9b0936754b7f25d6f43979666350621e0a59541868916f1b9b39de0e92dc2d108aec1177c9baea7858"
//	_pubkey, _ := hex.DecodeString(pubkeyStr)
//	WriteVlen(b, uint64(len(_pubkey)))
//	b.Write(_pubkey)
//
//	// _licperiod
//	//period := int64(1541648532)
//	binary.Write(b, binary.LittleEndian, int64(period))
//
//	// _licversion
//	//licversion := int32(1)
//	binary.Write(b, binary.LittleEndian, int32(licversion))
//
//	// fmt.Println(b.Len())
//	uint256 := NewSha2Hash(b.Bytes())
//	fmt.Println(uint256.String())
//
//	if uint256.String() != "cd77355ebc70f9d1b674a4d3dacfb09464dcfbf2b3055ab97ee8a7d28e447784" {
//		fmt.Println("failed")
//	}
//
//	// cd77355ebc70f9d1b674a4d3dacfb09464dcfbf2b3055ab97ee8a7d28e447784
//}

func Test_MakePrivateAddr(t *testing.T) {
	fmt.Println(MakePrivateAddr())
}

func Test_PublicKeyFromPrivateAddr(t *testing.T) {
	pub := PublicKeyFromPrivateAddr(privStr)
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
