package ca

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
		"testing"
	"net"
					)

var(
	txid = "829f9b963be709e8b138385fedb876990f892b99ca641401e8a88df531735ca7"
	voutid = uint32(1)

	period = int64(1543386972)
	licversion = int32(1)

	pubkeyStr = "0427989b89ebc4c8596a82830a877b5edbde0739ea8df6e7c8bbbdf3bf0857f555bb739f3e9ce9951a0f34be4ff4d7e771b544401139de4dcca4809a0aac4c6ed5"
	hashStr = "6b1b03d59c397fa77967a10e33259c58db295c1f8b66ba568a6c13ea0cc8ea21"

	signStr = "IEC1TEc+iLcJ2LsXFZ2Xdx7E8VmXuoW98unHqsh2uNXmBmMpI7Lm+/v1B7hrczJ0hO3pMh5yTgk/6aS7L/VnCGE="

	ucenterPubkeyStr = "03e947099921ee170da47a7acf48143c624d33950af362fc39a734b1b3188ec1e3"
)


func Test_MakeBTCHash(t *testing.T) {
	if MakeBTCHash(txid, voutid, pubkeyStr, period, licversion, ) !=hashStr {
		t.Failed()
	}
}

func Test_RequestSignature(t *testing.T) {
	type RequestMsg struct {
		size      int32
		version   int32
		timestamp int64
		questtype int32
		txid      string
		vountid   uint32
	}

	fmt.Println()

	msg := RequestMsg{
		size:      85,
		version:   111,
		timestamp: 0,
		questtype: 1,
		txid:      txid,
		vountid:   voutid,
	}

	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.BigEndian, msg.size)
	binary.Write(b, binary.LittleEndian, msg.version)
	binary.Write(b, binary.LittleEndian, msg.timestamp)
	binary.Write(b, binary.LittleEndian, msg.questtype)
	WriteVlen(b, uint64(len(msg.txid)))
	b.Write([]byte(msg.txid))
	binary.Write(b, binary.LittleEndian, msg.vountid)

	if b.Len()-4 != int(msg.size) {
		t.Fatal("error length")
	}

	conn, err := net.Dial("tcp", "132.232.98.139:5009")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Write(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	var size int32
	err = binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		t.Fatal(err)
	}

	// buf := make([]byte, size)
	// _, err = io.ReadAtLeast(conn, buf, int(size))
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// int             _msgversion; //111
	// int             _num;        //1
	// int             _nodetype;   //1

	// int          _version;
	// std::string  _txid;       //use
	// unsigned int _voutid;     //use
	// std::string  _privkey;    //
	// int          _status;     //
	// int          _licversion;  //use
	// int64_t      _licperiod;  //use
	// std::string  _licence;    //use
	// int64_t      _nodeperiod;

	type ResponseMsg struct {
		msgversion int32
		num        int32
		nodeType   int32

		version    int32
		txid       string
		voutid     uint32
		privkey    string
		status     int32
		licversion int32
		licperiod  int64
		license    string
		nodeperiod int64
	}

	res := &ResponseMsg{}

	content := conn

	// content := bytes.NewReader(buf)
	err = binary.Read(content, binary.LittleEndian, &res.msgversion)
	if err != nil {
		t.Fatal(err)
	}

	err = binary.Read(content, binary.LittleEndian, &res.num)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(content, binary.LittleEndian, &res.nodeType)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(content, binary.LittleEndian, &res.version)
	if err != nil {
		t.Fatal(err)
	}

	res.txid, err = ReadString(conn)
	if err != nil {
		t.Fatal(err)
	}

	err = binary.Read(content, binary.LittleEndian, &res.voutid)
	if err != nil {
		t.Fatal(err)
	}

	res.privkey, err = ReadString(conn)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(content, binary.LittleEndian, &res.status)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(content, binary.LittleEndian, &res.licversion)
	if err != nil {
		t.Fatal(err)
	}

	err = binary.Read(content, binary.LittleEndian, &res.licperiod)
	if err != nil {
		t.Fatal(err)
	}

	res.license, err = ReadString(conn)
	if err != nil {
		t.Fatal(err)
	}

	err = binary.Read(content, binary.LittleEndian, &res.nodeperiod)
	if err != nil {
		t.Fatal(err)
	}

	// verify
	if res.licperiod != period {
		t.Failed()
	}

	if res.voutid != voutid {
		t.Failed()
	}

	if res.license != signStr {
		t.Failed()
	}

	if res.licversion != licversion {
		t.Failed()
	}
}


func Test_VerifySignature(t *testing.T) {
	if !VerifySignature(hashStr, signStr, ucenterPubkeyStr) {
		t.Failed()
	}
}

func Test_Raw(t *testing.T) {
	b := bytes.NewBuffer(nil)
	// strMessageMagic
	strMessageMagic := "Ulord Signed Message:\n"
	WriteVlen(b, uint64(len(strMessageMagic)))
	b.Write([]byte(strMessageMagic))

	// _txid
	txid := "da29e46c4d805474cbed1e6771473c77b350858e990de26837fcbeb8a0271acf"
	WriteVlen(b, uint64(len(txid)))
	b.Write([]byte(txid))

	// _voutid
	voutid := uint32(0)
	binary.Write(b, binary.LittleEndian, uint32(voutid))

	// _pubkey
	pubkey := "041df129b64bf4091c6093b7c6253479b96541ce1fef817a9b0936754b7f25d6f43979666350621e0a59541868916f1b9b39de0e92dc2d108aec1177c9baea7858"
	_pubkey, _ := hex.DecodeString(pubkey)
	WriteVlen(b, uint64(len(_pubkey)))
	b.Write(_pubkey)

	// _licperiod
	licperiod := int64(1541648532)
	binary.Write(b, binary.LittleEndian, int64(licperiod))

	// _licversion
	licversion := int32(1)
	binary.Write(b, binary.LittleEndian, int32(licversion))

	// fmt.Println(b.Len())
	uint256 := NewSha2Hash(b.Bytes())
	fmt.Println(uint256.String())

	if uint256.String() != "cd77355ebc70f9d1b674a4d3dacfb09464dcfbf2b3055ab97ee8a7d28e447784" {
		fmt.Println("failed")
	}

	// cd77355ebc70f9d1b674a4d3dacfb09464dcfbf2b3055ab97ee8a7d28e447784
}


func Test_MakePrivateAddr(t *testing.T) {
	fmt.Println(MakePrivateAddr())
}

func Test_PublicKeyFromPrivateAddr(t *testing.T) {
	fmt.Println(PublicKeyFromPrivateAddr("5JDu9q8sWyc19kECjDtygeSzTYSACpVg6giyh7u7twus4B5cMCF"))
}