package service

import (
	"context"
	"fmt"
	"github.com/glvd/accipfs/dhcrypto"
	"net"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/glvd/accipfs"
	"github.com/glvd/accipfs/config"
	"github.com/glvd/accipfs/contract"
	"github.com/goextension/log"
	"github.com/ipfs/go-ipfs-http-client"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"go.uber.org/atomic"
)

const ipfsPath = ".ipfs"
const ipfsAPI = "api"

type nodeClientIPFS struct {
	lock *atomic.Bool
	cfg  config.Config
	api  *httpapi.HttpApi
	out  *color.Color
}

type nodeServerIPFS struct {
	name string
	cmd  *exec.Cmd
}

// PeerID ...
type PeerID struct {
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ID              string   `json:"ID"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	PublicKey       string   `json:"PublicKey"`
}

func newNodeIPFS(config config.Config) (*nodeClientIPFS, error) {

	return &nodeClientIPFS{
		cfg: config,
		//api:  api,
		lock: atomic.NewBool(false),
		out:  color.New(color.FgBlue),
	}, nil
}

// SwarmConnect ...
func (i *nodeClientIPFS) SwarmConnect(ctx context.Context, addr string) (e error) {
	ma, e := multiaddr.NewMultiaddr(addr)
	if e != nil {
		return e
	}
	info, e := peer.AddrInfoFromP2pAddr(ma)
	if e != nil {
		return e
	}
	e = i.api.Swarm().Connect(ctx, *info)
	if e != nil {
		return e
	}
	return nil
}

func (i *nodeClientIPFS) connect() (e error) {
	ma, err := multiaddr.NewMultiaddr(i.cfg.IPFS.Addr)
	if err != nil {
		return err
	}
	i.api, e = httpapi.NewApi(ma)
	return
}

// SwarmPeers ...
func (i *nodeClientIPFS) SwarmPeers(ctx context.Context) ([]iface.ConnectionInfo, error) {
	return i.api.Swarm().Peers(ctx)
}

// ID get self node info
func (i *nodeClientIPFS) ID(ctx context.Context) (pid *PeerID, e error) {
	pid = &PeerID{}
	e = i.api.Request("id").Exec(ctx, pid)
	if e != nil {
		return nil, e
	}
	return pid, nil
}

// PinAdd ...
func (i *nodeClientIPFS) PinAdd(ctx context.Context, hash string) (e error) {
	p := path.New(hash)
	return i.api.Pin().Add(ctx, p)
}

// PinLS ...
func (i *nodeClientIPFS) PinLS(ctx context.Context) (pins []iface.Pin, e error) {
	return i.api.Pin().Ls(ctx)
}

// PinRm ...
func (i *nodeClientIPFS) PinRm(ctx context.Context, hash string) (e error) {
	p := path.New(hash)
	return i.api.Pin().Rm(ctx, p)
}

// IsReady ...
func (i *nodeClientIPFS) IsReady() bool {
	api, e := httpapi.NewPathApi(filepath.Join(i.cfg.Path, ipfsPath))
	if e != nil {
		log.Errorw("new node ipfs", "error", e)
		return false
	}
	i.api = api
	return true
}

func (i *nodeClientIPFS) output(v ...interface{}) {
	fmt.Print(outputHead, " ")
	fmt.Print("[IPFS]", " ")
	_, _ = i.out.Println(v...)
}

// Run ...
func (i *nodeClientIPFS) Run() {
	i.output("syncing node")
	if i.lock.Load() {
		i.output("ipfs node is already running")
		return
	}
	i.lock.Store(true)
	defer i.lock.Store(false)
	i.output("ipfs sync running")
	if !i.IsReady() {
		i.output("waiting for ready")
		return
	}
	//// get self node info
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Duration(i.cfg.IPFS.Timeout)*time.Second)
	defer cancelFunc()
	pid, e := i.ID(timeout)
	if e != nil {
		log.Errorw("run get id", "tag", outputHead, "error", e)
		return
	}

	nid := pid.ID
	i.output("id", nid)
	// get ipfs swarm nodes
	var peers []string
	var publicNodes []string
	timeout2, cancelFunc2 := context.WithTimeout(context.Background(), time.Duration(i.cfg.IPFS.Timeout)*time.Second)
	defer cancelFunc2()
	infos, e := i.SwarmPeers(timeout2)
	if e != nil {
		log.Errorw(outputHead, "tag", "run get peers", "error", e)
		return
	}
	for _, info := range infos {
		i.output("peers", info.ID().String(), "ip", info.Address())
		conn, err := manet.Dial(info.Address())
		// p2p proxy node
		if err != nil {
			//TODO:
			//ipfsAddr := "/ipfs/" + nodeID + "/p2p-circuit/ipfs/" + .Peer
			//peers = append(peers, ipfsAddr)
		} else {
			ipfsAddr := info.Address().String()
			publicNodes = append(publicNodes, ipfsAddr)
			conn.Close()
		}
	}
	//i.output("[当前IPFS总节点数]", len(peers)+len(publicNodes))
	i.output("exists IPFS nodes:", len(publicNodes))
	// get nodes info
	//if len(peers) == 0 {
	//	fmt.Println("<IPFS节点状态已是最新>")
	//	return
	//}
	cl := contract.Loader()
	ac, auth, client := cl.AccelerateNode()
	defer client.Close()

	cPeers, err := ac.GetIpfsNodes(nil)
	if err != nil {
		log.Errorw("ipfs node", "error", err)
		return
	}
	cNodes, err := ac.GetPublicIpfsNodes(nil)
	if err != nil {
		log.Errorw("public ipfs node", "error", err)
		return
	}
	cPeers = i.decodeNodes(cPeers)
	cNodes = i.decodeNodes(cNodes)

	//TODO:fix sta
	fmt.Println("[adding ipfs nodes]", difference(peers, cPeers))
	fmt.Println("[adding public ipfs nodes]", difference(publicNodes, cNodes))
	//// delete nodes
	var deleteIdx []int
	for _, dNode := range difference(cNodes, getAccessibleIpfsNodes(cNodes, "4001")) {
		for idx, cNode := range cNodes {
			if cNode == dNode {
				deleteIdx = append(deleteIdx, idx)
			}
		}
	}
	//TODO:fix end
	if len(deleteIdx) > 0 {
		var err error
		sort.Sort(sort.Reverse(sort.IntSlice(deleteIdx)))
		for _, idx := range deleteIdx {
			_, err = ac.DeletePublicIpfsNodes(auth, uint32(idx))
		}

		if err != nil {
			fmt.Println("<删除失效节点失败>", err.Error())
		} else {
			fmt.Println("[删除失效节点成功]")
		}
	}

	// add new nodes
	for _, n := range i.encodeNodes(difference(peers, cPeers)) {
		if n == "" {
			continue
		}
		_, err = ac.AddIpfsNodes(auth, []string{n})
	}
	for _, n := range i.encodeNodes(difference(publicNodes, cNodes)) {
		if n == "" {
			continue
		}
		_, err = ac.AddPublicIpfsNodes(auth, []string{n})
	}

	if err != nil {
		fmt.Println("[添加节点失败]", err.Error())
	} else {
		fmt.Println("[添加节点成功] ")
	}
	i.output("<IPFS同步完成>")
	return
}

// Start ...
func (n *nodeServerIPFS) Start() {
	fmt.Println("starting", n.name)
}

// NodeServerIPFS ...
func NodeServerIPFS(cfg config.Config) Node {
	cmd := exec.Command(cfg.IPFS.Name, "")
	cmd.Env = accipfs.Environ()
	return &nodeServerIPFS{cmd: cmd}
}

// Address ...
type Address struct {
	Addr string
	Port string
}

// StringToAddr ...
func StringToAddr(s string) (a *Address, e error) {
	return
	//// consume trailing slashes
	//s = strings.TrimRight(s, "/")
	//
	//var b bytes.Buffer
	//sp := strings.Split(s, "/")
	//
	//if sp[0] != "" {
	//	return nil, fmt.Errorf("failed to parse multiaddr %q: must begin with /", s)
	//}
	//
	//// consume first empty elem
	//sp = sp[1:]
	//
	//if len(sp) == 0 {
	//	return nil, fmt.Errorf("failed to parse multiaddr %q: empty multiaddr", s)
	//}
	//
	//for len(sp) > 0 {
	//	log.Infow("sp info", "tag", outputHead, "info", sp)
	//	name := sp[0]
	//	p := multiaddr.ProtocolWithName(name)
	//	if p.Code == 0 {
	//		return nil, fmt.Errorf("failed to parse multiaddr %q: unknown protocol %s", s, sp[0])
	//	}
	//
	//	_, _ = b.Write(multiaddr.CodeToVarint(p.Code))
	//	log.Info(sp)
	//	sp = sp[1:]
	//	if p.Size == 0 { // no length.
	//		continue
	//	}
	//
	//	if len(sp) < 1 {
	//		return nil, fmt.Errorf("failed to parse multiaddr %q: unexpected end of multiaddr", s)
	//	}
	//
	//	if p.Path {
	//		// it's a path protocol (terminal).
	//		// consume the rest of the address as the next component.
	//		sp = []string{"/" + strings.Join(sp, "/")}
	//	}
	//
	//	a, err := p.Transcoder.StringToBytes(sp[0])
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to parse multiaddr %q: invalid value %q for protocol %s: %s", s, sp[0], p.Name, err)
	//	}
	//	if p.Size < 0 { // varint size.
	//		_, _ = b.Write(multiaddr.CodeToVarint(len(a)))
	//	}
	//	b.Write(a)
	//	log.Info(sp)
	//	sp = sp[1:]
	//}
	//
	//return b.Bytes(), nil
}

func (i *nodeClientIPFS) decodeNodes(nodes []string) []string {
	// init contract
	var decodeNodes []string
	decoder := dhcrypto.NewCipherDecode([]byte(i.cfg.PrivateKey), dateKey)
	if len(nodes) == 0 {
		return decodeNodes
	}
	for _, node := range nodes {
		decoded, err := decoder.Decode(node)
		if err != nil {
			continue
		}
		decodeNodes = append(decodeNodes, string(decoded))
	}
	return decodeNodes
}

func (i *nodeClientIPFS) encodeNodes(nodes []string) []string {
	var encodedNodes []string
	encoder := dhcrypto.NewCipherEncoder([]byte(i.cfg.PublicKey), 10, dateKey)
	for _, node := range nodes {
		encoded, err := encoder.Encode(node)
		if err != nil {
			continue
		}
		encodedNodes = append(encodedNodes, string(encoded))
	}
	return encodedNodes
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func getAccessibleIpfsNodes(addresses []string, port string) []string {
	var accessible []string
	for _, address := range addresses {
		strs := strings.Split(address, "/ip4/")
		if len(strs) < 2 {
			continue
		}
		ip := strings.Split(strs[1], "/")[0]
		if len(ip) < 8 {
			continue
		}
		conn, err := net.DialTimeout("tcp", ip+":"+port, 3*time.Second)
		if err == nil {
			addr := strs[0] + "@" + ip + ":" + port
			accessible = append(accessible, addr)
			conn.Close()
		} else {
			fmt.Println("[dial err]", err)
		}

	}
	return accessible
}
