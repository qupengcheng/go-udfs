package verify

import (
	"context"

	"github.com/pkg/errors"

	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"

	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"

	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"

	"time"

	msmux "gx/ipfs/QmbXRda5H2K3MSQyWWxTMtd8DWuguEBUCe6hpxfXVpFUGj/go-multistream"

	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/udfs/ca"
	pb "github.com/ipfs/go-ipfs/udfs/go-libp2p/p2p/protocol/verify/pb"
)

const ProtocolVerify = "/ipfs/verify/0.0.1"

var log = logging.Logger("net/verify")

func readVerifyAskValue(s inet.Stream) (string, error) {
	r := ggio.NewDelimitedReader(s, 1024)
	mes := pb.VerifyAsk{}
	if err := r.ReadMsg(&mes); err != nil {
		return "", errors.Wrap(err, "read verify ask message failed")
	}

	return mes.GetRandom(), nil
}

func sendVerifyAskMsg(s inet.Stream) (string, error) {

	rhash := ca.MakeRandomHash()

	w := ggio.NewDelimitedWriter(s)
	mes := pb.VerifyAsk{
		Random: &rhash,
	}
	if err := w.WriteMsg(&mes); err != nil {
		return "", errors.Wrap(err, "write verify ask message failed")
	}

	return rhash, nil
}

func doVerify(mes *pb.Verify, rhash string) error {
	// verify license
	nodehash := ca.MakeNodeInfoHash(mes.GetTxid(), mes.GetVoutid(), mes.GetPubkey(), mes.GetPeriod(), mes.GetLicversion())
	if nodehash == "" {
		return errors.New("make node hash failed")
	}

	ok, err := ca.VerifySignature(nodehash, mes.GetLicense(), ca.LicenseCenterPubkey())
	if err != nil {
		return errors.Wrap(err, "verify license error")
	}

	if !ok {
		return errors.New("verify license failed")
	}

	// verify node signature
	ok, err = ca.VerifySignature(rhash, mes.GetRandomSign(), mes.GetPubkey())
	if err != nil {
		return errors.Wrap(err, "verify node signature error")
	}

	if !ok {
		return errors.New("verify node signature failed")
	}

	return nil
}

// VerifyConn verify whether this connection is legal, if not, must close this connection
func VerifyConn(c inet.Conn) error {

	// make new stream
	s, err := c.NewStream()
	if err != nil {
		log.Debugf("error opening initial stream for %s: %s", ProtocolVerify, err)
		log.Event(context.TODO(), "VerifyConnFailed", c.RemotePeer())
		return errors.Wrap(err, "new stream failed")
	}
	defer inet.FullClose(s)

	s.SetProtocol(ProtocolVerify)

	err = s.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return errors.Wrap(err, "set read deadline failed")
	}
	err = s.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return errors.Wrap(err, "set write deadline failed")
	}

	// ok give the response to our handler.
	if err := msmux.SelectProtoOrFail(ProtocolVerify, s); err != nil {
		log.Event(context.TODO(), "VerifyConnFailed", c.RemotePeer(), logging.Metadata{"error": err})
		return errors.Wrap(err, "SelectProtoOrFail for verify failed")
	}

	// send verify ask msg
	rhash, err := sendVerifyAskMsg(s)
	if err != nil {
		return errors.Wrap(err, "send verify ask msg failed")
	}

	// read verify info
	r := ggio.NewDelimitedReader(s, 2048)
	mes := &pb.Verify{}
	if err := r.ReadMsg(mes); err != nil {
		return errors.Wrap(err, "error reading verify message")
	}

	log.Debugf("%s received message from %s %s", ProtocolVerify,
		c.RemotePeer(), c.RemoteMultiaddr())

	err = doVerify(mes, rhash)
	if err != nil {
		return err
	}

	return nil
}

func checkVerifyInfo(vfi *config.VerifyInfo) error {
	if len(vfi.Txid) == 0 {
		return errors.New("the field <txid> in verify info is empty")
	}

	if len(vfi.Secret) == 0 {
		return errors.New("the field <secret> in verify info is empty")
	}

	return nil
}

func RequestHandler(s inet.Stream, h host.Host) {
	defer s.Close()

	// read verify ask msg
	rhash, err := readVerifyAskValue(s)
	if err != nil {
		log.Error("read verify ask value failed: ", err)
		return
	}

	err = requestVerify(s, h, rhash)
	if err != nil {
		log.Error("request verify failed: ", err)
		return
	}
}

func requestVerify(s inet.Stream, h host.Host, rhash string) error {

	// get the verify info
	repoInf, err := h.Peerstore().Get(s.Conn().LocalPeer(), "repo")
	if err != nil {
		return errors.Wrap(err, "got repo object from peerstore failed")
	}
	repo := repoInf.(repo.Repo)
	cfg, err := repo.Config()
	if err != nil {
		return errors.Wrap(err, "got config object from repo failed")
	}
	vfi := &cfg.Verify

	err = checkVerifyInfo(vfi)
	if err != nil {
		return errors.Wrap(err, "check local verify info failed")
	}

	// make signature for random hash
	sign, err := ca.Sign(rhash, vfi.Secret)
	if err != nil {
		return errors.Wrap(err, "sign for rhash failed")
	}

	// request license
	vfi.Pubkey = ca.PublicKeyFromPrivateAddr(vfi.Secret)
	lbi, err := ca.RequestLicense(vfi.Txid, vfi.Voutid)
	if err != nil {
		return errors.Wrap(err, "request license failed")
	}
	vfi.License = lbi.License
	vfi.Period = lbi.LicPeriod
	vfi.Licversion = lbi.Licversion

	// send verify msg
	w := ggio.NewDelimitedWriter(s)
	mes := pb.Verify{
		Txid:       &vfi.Txid,
		Voutid:     &vfi.Voutid,
		Period:     &vfi.Period,
		Pubkey:     &vfi.Pubkey,
		License:    &vfi.License,
		Licversion: &vfi.Licversion,
		RandomSign: &sign,
	}
	w.WriteMsg(&mes)

	return nil
}
