package netw

import (
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/pool"
	"github.com/Centny/gwf/util"
	"net"
	"time"
)

//the client connection pool.
type NConPool struct {
	*LConPool //base connection pool.
}

//new client connection pool.
func NewNConPool(p *pool.BytePool, h CCHandler, n string) *NConPool {
	return &NConPool{
		LConPool: NewLConPoolV(p, h, n, NewConH),
	}
}
func NewNConPool2(p *pool.BytePool, h CCHandler) *NConPool {
	return NewNConPool(p, h, "C-")
}

//dail one connection.
func (n *NConPool) Dail(addr string) (Con, error) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	cc := n.NewCon(n, n.P, con)
	if !n.H.OnConn(cc) {
		return nil, util.Err("OnConn return false for %v", addr)
	}
	n.RunC(cc)
	return cc, nil
}

func Dail(p *pool.BytePool, addr string, h CCHandler) (*NConPool, Con, error) {
	return DailN(p, addr, h, NewCon)
}
func DailN(p *pool.BytePool, addr string, h CCHandler, ncf NewConF) (*NConPool, Con, error) {
	nc := NewNConPool2(p, h)
	nc.NewCon = ncf
	cc, err := nc.Dail(addr)
	return nc, cc, err
}

type NConRunner struct {
	*NConPool
	C         Con
	ConH      ConHandler
	Connected bool
	Running   bool
	Retry     time.Duration
	Tick      time.Duration
	TickData  []byte
	//
	Addr    string
	NCF     NewConF
	BP      *pool.BytePool
	CmdH    CmdHandler
	ShowLog bool //setting the ShowLog to Con_
	TickLog bool //if show the tick log.
	WC      chan int
}

func (n *NConRunner) OnConn(c Con) bool {
	if n.ConH == nil {
		return true
	}
	return n.ConH.OnConn(c)
}
func (n *NConRunner) OnClose(c Con) {
	if n.ConH != nil {
		n.ConH.OnClose(c)
	}
	if n.Running {
		go n.Try()
	}
}
func (n *NConRunner) StartRunner() {
	go n.Try()
	go n.StartTick()
	log.D("starting runner...")
}
func (n *NConRunner) StopRunner() {
	n.Running = false
	if n.NConPool != nil {
		n.NConPool.Close()
	}
	log.D("stopping runner...")
	n.WC <- 0
}
func (n *NConRunner) StartTick() {
	if len(n.TickData) < 1 {
		return
	}
	go n.RunTick_()
}
func (n *NConRunner) write_tick() {
	c := n.C
	if c == nil {
		log.D("sending tick message err: the connection is nil")
		return
	}
	_, err := c.Writeb(n.TickData)
	if err != nil {
		log.W("send tck message err:%v", err)
		return
	}
	if n.TickLog {
		log.D("sending tick message to Push Server")
	}
}
func (n *NConRunner) RunTick_() {
	tk := time.Tick(n.Tick * time.Millisecond)
	n.Running = true
	log.I("starting tick(%vms) to server(%v)", int(n.Tick), n.Addr)
	for n.Running {
		select {
		case <-tk:
			n.write_tick()
		}
	}
	log.I("tick to server(%v) will stop", n.Addr)
}
func (n *NConRunner) Try() {
	n.Running = true
	for n.Running {
		err := n.Dail()
		log.D("connect to server(%v) success", n.Addr)
		if err == nil {
			break
		}
		log.D("try connect to server(%v) err:%v,will retry after %v ms", n.Addr, err.Error(), n.Retry)
		time.Sleep(n.Retry * time.Millisecond)
	}
	// log.D("connect try stopped")
}
func (n *NConRunner) Dail() error {
	n.Connected = false
	nc, cc, err := DailN(n.BP, n.Addr, NewCCH(n, n.CmdH), n.NCF)
	if err != nil {
		return err
	}
	n.NConPool = nc
	n.C = cc
	n.Connected = true
	cc.(*Con_).ShowLog = n.ShowLog
	return nil
}

func NewNConRunnerN(bp *pool.BytePool, addr string, h CmdHandler, ncf NewConF) *NConRunner {
	return &NConRunner{
		Addr:     addr,
		NCF:      ncf,
		BP:       bp,
		CmdH:     h,
		Retry:    5000,
		Tick:     30000,
		TickData: []byte("Tick\n"),
		WC:       make(chan int),
	}
}
func NewNConRunner(bp *pool.BytePool, addr string, h CmdHandler) *NConRunner {
	return NewNConRunnerN(bp, addr, h, NewCon)
}
