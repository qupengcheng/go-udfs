package ca

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"crypto/rand"

	"github.com/pkg/errors"
)

type BTCEncoder struct {
	buf *bytes.Buffer
}

func NewBTCEncoder(olds ...[]byte) *BTCEncoder {
	buf := bytes.NewBuffer(nil)
	for _, old := range olds {
		buf.Write(old)
	}

	return &BTCEncoder{
		buf: buf,
	}
}

// Write write date to NewSignatureMaker for make license, now only support number、 string and []byte type.
func (t *BTCEncoder) Write(data interface{}) error {
	switch data.(type) {
	case string:
		WriteVlen(t.buf, uint64(len(data.(string))))
		t.buf.WriteString(data.(string))
	case []byte:
		t.buf.Write(data.([]byte))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		binary.Write(t.buf, binary.LittleEndian, data)
	default:
		return errors.Errorf("unsupport data type %v to write", reflect.TypeOf(data))
	}
	return nil
}

func (t *BTCEncoder) Bytes() []byte {
	return t.buf.Bytes()
}

func (t *BTCEncoder) Len() int {
	return t.buf.Len()
}

func (t *BTCEncoder) Hash() string {
	uint256 := NewSha2Hash(t.buf.Bytes())
	return uint256.String()
}

func MakeBTCHash(txid string, voutid uint32, pubkey string, licperiod int64, licversion int32) string {
	lm := NewBTCEncoder(nil)
	lm.Write("Ulord Signed Message:\n")
	lm.Write(txid)
	lm.Write(voutid)
	_pubkey, _ := hex.DecodeString(pubkey)
	lm.Write(string(_pubkey))
	lm.Write(licperiod)
	lm.Write(licversion)
	return lm.Hash()
}

// func MakeSignature(txid string, voutid uint32, pubkey string, licperiod int64, licversion int32) string {
// 	b := bytes.NewBuffer(nil)
// 	// strMessageMagic
// 	strMessageMagic := "Ulord Signed Message:\n"
// 	WriteVlen(b, uint64(len(strMessageMagic)))
// 	b.Write([]byte(strMessageMagic))

// 	WriteVlen(b, uint64(len(txid)))
// 	b.Write([]byte(txid))

// 	// _voutid
// 	binary.Write(b, binary.LittleEndian, voutid)

// 	// _pubkey
// 	_pubkey, _ := hex.DecodeString(pubkey)
// 	WriteVlen(b, uint64(len(_pubkey)))
// 	b.Write(_pubkey)

// 	// _licperiod
// 	binary.Write(b, binary.LittleEndian, licperiod)

// 	// _licversion
// 	binary.Write(b, binary.LittleEndian, licversion)

// 	// fmt.Println(b.Len())
// 	return NewSha2Hash(b.Bytes()).String()
// }

//if compressed {
//res = make([]byte, 33)
//} else {
//res = make([]byte, 65)
//}
//
//if !secp256k1.BaseMultiply(priv_key, res) {
//res = nil
//}



func MakePrivateAddr() string{
	key := make([]byte, 32)
	rand.Read(key)
	addr := NewPrivateAddr(key, 128, false)
	//fmt.Println(
	return addr.String()
}


func PublicKeyFromPrivateAddr(privateAddr string) string{
	addr, err := DecodePrivateAddr(privateAddr)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(addr.Pubkey)
}

func VerifySignature(hash, licenseInBase64, pubkeyInHex string) bool {

	_, sign, err := ParseMessageSignature(licenseInBase64)
	if err != nil {
		return false
	}

	ucenterPubkeyBytes, err := hex.DecodeString(pubkeyInHex)
	if err != nil {
		return false
	}
	res, err := NewPublicKey(ucenterPubkeyBytes)
	if err != nil {
		return false
	}


	return EcdsaVerify(res.Bytes(true), sign.Bytes(), NewUint256FromString(hash).Bytes())
}


//
//
//// 加密
//func RsaEncryptWithPublicKey(data , key []byte) ([]byte, error) {
//	block, _ := pem.Decode(key) //将密钥解析成公钥实例
//	if block == nil {
//		return nil, errors.New("key error")
//	}
//	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes) //解析pem.Decode（）返回的Block指针实例
//	if err != nil {
//		return nil, err
//	}
//	pub := pubInterface.(*rsa.PublicKey)
//	return rsa.EncryptPKCS1v15(rand.Reader, pub, data) //RSA算法加密
//}
//
//
//// 解密
//func RsaDecryptWithPrivateKey(data , key []byte) ([]byte, error) {
//	block, _ := pem.Decode(key) //将密钥解析成私钥实例
//	if block == nil {
//		return nil, errors.New("private key error!")
//	}
//	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes) //解析pem.Decode（）返回的Block指针实例
//	if err != nil {
//		return nil, err
//	}
//	return rsa.DecryptPKCS1v15(rand.Reader, priv, data) //RSA算法解密
//}
