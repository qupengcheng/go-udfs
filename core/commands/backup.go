package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"

	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit"

	cid "gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"

	cmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"
	e "github.com/ipfs/go-ipfs/core/commands/e"
	config "github.com/ipfs/go-ipfs/repo/config"
)

const ProtocolBackup protocol.ID = "/backup/0.0.1"
const numberForBackup int = 2
const timeoutForLookup = 1 * time.Minute

type BackupResult struct {
	ID  string
	Msg string `json:",omitempty"`
}

type BackupOutput struct {
	Success []*BackupResult `json:",omitempty"`
	Failed  []*BackupResult `json:",omitempty"`
}

var BackupCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "Backup objects to remote node storage.",
		ShortDescription: "Stores an IPFS object(s) from a given path locally to remote disk.",
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("ipfs-path", true, false, "Path to object(s) to be pinned.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		// get cid
		c, err := cid.Decode(req.Arguments()[0])
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		// get peers for backup

		toctx, _ := context.WithTimeout(n.Context(), timeoutForLookup)
		closestPeers, err := n.DHT.GetClosestPeers(toctx, c.KeyString())
		if err != nil {
			res.SetError(errors.Wrap(err, "got closest peers timeout"), cmdkit.ErrNormal)
			return
		}

		peers := make(map[peer.ID]struct{}, 0)
		lookup := true
		for lookup {
			select {
			case p, closed := <-closestPeers:
				if closed {
					lookup = false
				}

				// issue: it seems get a empty id sometimes ?
				if p.Pretty() == "" {
					log.Error("BackupCmd got a empty closest peer!")
				} else {
					peers[p] = struct{}{}
					if len(peers) >= numberForBackup {
						lookup = false
					}
				}
			}
		}

		if len(peers) < numberForBackup {
			res.SetError(errors.Errorf("Failed to find the minimum number of closest peers required: %d/%d", len(peers),
				numberForBackup), cmdkit.ErrNormal)
			return
		}

		log.Debug("found the nodes to backup")
		peersForBackup := peers

		//peers, err := loadBootstrapPeers(n)
		//if err != nil {
		//	res.SetError(errors.New("failed to parse bootstrap peers from config"), cmdkit.ErrNormal)
		//	return
		//}
		//
		//var connectedPeers []pstore.PeerInfo
		//for _, p := range peers {
		//	if n.PeerHost.Network().Connectedness(p.ID) == inet.Connected {
		//		connectedPeers = append(connectedPeers, p)
		//	}
		//}
		//if len(connectedPeers) < numberForBackup {
		//	res.SetError(errors.New("not enught bootstrap node to backup"), cmdkit.ErrNormal)
		//	return
		//}

		// random some node for backup
		//var peersForBackup []pstore.PeerInfo
		//if len(connectedPeers) > numberForBackup {
		//	// get closest peers for backup
		//
		//
		//	// get random peers for backup
		//	for _, val := range rand.Perm(len(connectedPeers)) {
		//		peersForBackup = append(peersForBackup, connectedPeers[val])
		//		if len(peersForBackup) >= numberForBackup {
		//			break
		//		}
		//	}
		//}else{
		//	peersForBackup = connectedPeers
		//}

		// 发送cid
		results := make(chan *BackupResult, len(peersForBackup))
		var wg sync.WaitGroup
		for p := range peersForBackup {
			wg.Add(1)
			go func(id peer.ID) {
				e := doBackup(n, id, c)
				if e != nil {
					results <- &BackupResult{
						ID:  id.Pretty(),
						Msg: e.Error(),
					}
				} else {
					results <- &BackupResult{
						ID: id.Pretty(),
					}
				}
				wg.Done()
			}(p)
		}
		go func() {
			wg.Wait()
			close(results)
		}()

		output := &BackupOutput{}
		for r := range results {
			if r.Msg != "" {
				output.Failed = append(output.Failed, r)
			} else {
				output.Success = append(output.Success, r)
			}
		}

		res.SetOutput(output)
	},
	Type: BackupOutput{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			out, ok := v.(*BackupOutput)
			if !ok {
				return nil, e.TypeErr(out, v)
			}

			buf := new(bytes.Buffer)
			for _, s := range out.Success {
				fmt.Fprintf(buf, "backup success to %s\n", s.ID)
			}
			for _, f := range out.Failed {
				fmt.Println(f.ID, f.Msg)
				fmt.Fprintf(buf, "backup failed to %s : %s\n", f.ID, f.Msg)
			}

			return buf, nil
		},
	},
}

func doBackup(n *core.IpfsNode, id peer.ID, c *cid.Cid) error {

	s, err := n.PeerHost.NewStream(n.Context(), id, ProtocolBackup)
	if err != nil {
		return err
	}
	defer s.Close()

	// TODO: consider to use protobuf, now just direct send the cid
	_, err = s.Write([]byte(c.KeyString() + "\n"))
	if err != nil {
		return err
	}

	// read result
	buf := bufio.NewReader(s)
	bs, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	if bs == "\n" {
		log.Debugf("backup %s successd to node %s\n", c.String(), id.Pretty())
		return nil
	}

	return errors.New(bs)
}

func pidsToStrings(ps []peer.ID) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Pretty())
	}
	return out
}

func loadBootstrapPeers(n *core.IpfsNode) ([]pstore.PeerInfo, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	parsed, err := cfg.BootstrapPeers()
	if err != nil {
		return nil, err
	}
	return toPeerInfos(parsed), nil
}

func toPeerInfos(bpeers []config.BootstrapPeer) []pstore.PeerInfo {
	pinfos := make(map[peer.ID]*pstore.PeerInfo)
	for _, bootstrap := range bpeers {
		pinfo, ok := pinfos[bootstrap.ID()]
		if !ok {
			pinfo = new(pstore.PeerInfo)
			pinfos[bootstrap.ID()] = pinfo
			pinfo.ID = bootstrap.ID()
		}

		pinfo.Addrs = append(pinfo.Addrs, bootstrap.Transport())
	}

	var peers []pstore.PeerInfo
	for _, pinfo := range pinfos {
		peers = append(peers, *pinfo)
	}

	return peers
}

func SetupBackupHandler(node *core.IpfsNode) {
	node.PeerHost.SetStreamHandler(ProtocolBackup, func(s inet.Stream) {
		var errRet error
		defer func() {
			var e error
			if errRet != nil {
				log.Error("backup-hander failed: ", errRet.Error())
				_, e = s.Write([]byte(errRet.Error() + "\n"))
			} else {
				_, e = s.Write([]byte("\n"))
			}

			if e != nil {
				log.Error("backup-handler send result failed: ", e.Error())
			} else {
				log.Debug("backup-handler send result success")
			}

			s.Close()
		}()

		select {
		case <-node.Context().Done():
			return
		default:
		}

		log.Debug("backup-handler receive request from", s.Conn().RemoteMultiaddr().String(), "/", s.Conn().RemotePeer().Pretty())

		// TODO: consider use protobuf, now just direct get cid
		buf := bufio.NewReader(s)
		bs, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			errRet = errors.Wrap(err, "backup-handler read bytes failed")
			return
		}

		c, err := cid.Cast([]byte(bs[:len(bs)-1]))
		if err != nil {
			errRet = errors.Wrap(err, "parse cid failed")
			return
		}
		log.Debug("backup-handler cid=", c.String())

		// do pin add
		co, err := exec.Command("ipfs", "pin", "add", c.String()).CombinedOutput()
		if err != nil {
			errRet = errors.Wrapf(err, "backup-handler run pin command for %s failed", c.String())
			return
		}

		log.Debug("backup-handler run pin command result: ", string(co))
	})
}
