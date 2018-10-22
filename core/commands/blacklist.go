package commands

import (
	"encoding/csv"

	"context"

	"io"

	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"

	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit"

	"time"

	"fmt"

	core "github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"github.com/ipfs/go-ipfs/core/coreunix"
	path "github.com/ipfs/go-ipfs/path"
	"github.com/pkg/errors"
)

const blacklistFile = "/ipns/QmbETUnWes7zdwZkkMGgPRtpZAYpFPxrUrCYy7fWi7JjFY/blacklist"

var BlacklistCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "Run blacklist service.",
		ShortDescription: "run a blacklist refresh operation right now.",
	},

	Arguments: []cmdkit.Argument{},
	Options:   []cmdkit.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		node, err := GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		if !node.OnlineMode() {
			if err := node.SetupOfflineRouting(); err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}
		}

		err = refreshBlacklist(req.Context, node, 1)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

	},
}

func refreshBlacklist(ctx context.Context, n *core.IpfsNode, minFailed int) error {
	dagReader, err := coreunix.Cat(ctx, n, blacklistFile)
	if err != nil {
		return errors.Errorf("read blacklist failed: %v\n", err.Error())
	}

	csvReader := csv.NewReader(dagReader)
	failedCount := 0
	for {
		record, err := csvReader.Read()
		if record == nil || err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Errorf("read record from blacklist failed: %v\n", err.Error())

		}
		log.Debug("blacklist record:", record)

		err = handleBlacklistRecord(ctx, n, record)
		if err != nil {
			failedCount++
			if minFailed > 0 && failedCount >= minFailed {
				return err
			}

			log.Error(err)
		}
	}

}

func RunBlacklistRefreshService(ctx context.Context, n *core.IpfsNode) error {
	conf, err := n.Repo.Config()
	if err != nil {
		return errors.Wrap(err, "got config failed")
	}

	d := 10 * time.Second
	if conf.Blacklist.Interval != "" {
		d, err = time.ParseDuration(conf.Blacklist.Interval)
		if err != nil {
			return errors.Wrap(err, "parse config.Blacklist.Interval failed")
		}
	}

	tm := time.NewTimer(d)

	go func() {
		defer tm.Stop()

		for {
			select {
			case <-tm.C:
				refreshBlacklist(ctx, n, -1)

				tm.Reset(d)

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil

}

func handleBlacklistRecord(ctx context.Context, n *core.IpfsNode, record []string) error {
	c, err := pathToCid(record[0])
	if err != nil {
		return errors.Errorf("got blacklist record cid failed from %v: %v\n", record, err.Error())
	}

	has, err := n.Blockstore.Has(c)
	if err != nil {
		return errors.Errorf("check blacklist record cid %s exist failed: %v\n", c.String(), err.Error())
	}

	if !has {
		return nil
	}

	_, pined, err := n.Pinning.IsPinned(c)
	if err != nil {
		return errors.Errorf("check blacklist record cid %s pined failed: %v\n", c.String(), err.Error())
	}

	if pined {
		_, err := corerepo.Unpin(n, ctx, record[:1], true)
		if err != nil {
			return errors.Errorf("unpin blacklist record cid %s failed: %v\n", c.String(), err.Error())
		}
		fmt.Println("unpin ", record[0])
	}

	err = corerepo.Remove(n, ctx, []*cid.Cid{c}, true, false)
	if err != nil {
		return errors.Errorf("blacklist record cid %s remove from repo failed: %v\n", c.String(), err.Error())
	}

	fmt.Println("remove ", c.String())

	return nil
}

func pathToCid(pstr string) (*cid.Cid, error) {
	p, err := path.ParsePath(pstr)
	if err != nil {
		return nil, err
	}

	c, _, err := path.SplitAbsPath(p)
	if err != nil {
		return nil, err
	}

	return c, nil
}
