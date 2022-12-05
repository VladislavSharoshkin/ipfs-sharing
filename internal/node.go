package internal

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/corerepo"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo/fsrepo"
	"io"
	"os"
	"path/filepath"
)

type Node struct {
	IpfsNode *core.IpfsNode
	CoreAPI  icore.CoreAPI
}

func NewNode(ctx context.Context, repoPath string) (*Node, error) {
	err := setupPlugins(repoPath)
	if err != nil {
		return nil, err
	}

	if !fsrepo.IsInitialized(repoPath) {
		conf, err := config.Init(io.Discard, 2048)
		if err != nil {
			return nil, err
		}
		conf.Experimental.FilestoreEnabled = true
		err = fsrepo.Init(repoPath, conf)
		if err != nil {
			return nil, err
		}
	}
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTClientOption,
		Repo:    repo,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	nodeApi, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, err
	}

	return &Node{node, nodeApi}, nil
}

func setupPlugins(path string) error {
	plugins, err := loader.NewPluginLoader(filepath.Join(path, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func (nod *Node) Upload(dir string) (string, error) {

	fileInfo, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	sf, err := files.NewSerialFile(dir, false, fileInfo)
	if err != nil {
		return "", err
	}

	opts := []options.UnixfsAddOption{
		options.Unixfs.Pin(true),
		options.Unixfs.CidVersion(1),
		options.Unixfs.RawLeaves(true),
		options.Unixfs.Nocopy(true),
	}

	add, err := nod.CoreAPI.Unixfs().Add(context.Background(), sf, opts...)
	if err != nil {
		return "", err
	}

	return add.Cid().String(), nil
}

func (nod *Node) Delete(cid cid.Cid) error {
	return nod.IpfsNode.Filestore.DeleteBlock(context.Background(), cid)
}

func (nod *Node) GC() (err error) {
	corerepo.GarbageCollectAsync(nod.IpfsNode, context.Background())

	return
}
