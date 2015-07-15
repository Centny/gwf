package rc

import (
	"fmt"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/netw"
	"github.com/Centny/gwf/netw/impl"
	"github.com/Centny/gwf/pool"
	"github.com/Centny/gwf/util"
	"runtime"
	"testing"
	"time"
)

////////////////////////////////////
//////// server handler ////////////
////////////////////////////////////
type rc_s_h struct {
	L *RC_Listener_m
}

func (r *rc_s_h) OnCmd(c netw.Cmd) int {
	var args util.Map
	_, err := c.V(&args)
	if err != nil {
		log.E("rc_s_h V error:%v", err.Error())
		return -1
	}
	var cmd, name, msg string
	err = args.ValidF(`
		c,R|S,L:0;
		n,O|S,L:0;
		m,O|S,L:0;
		`, &cmd, &name, &msg)
	if err != nil {
		log.E("rc_s_h valid args error:%v", err.Error())
		return -1
	}
	switch cmd {
	case "l":
		//login by message connection.(usage 1)
		if len(name) < 1 {
			log.E("login name is empty")
			return -1
		}
		r.L.AddC_c(name, c)
		log.D("S(m)->login success by name(%v)", name)
		return 0
	case "m":
		log.D("S(m)->receive message(%v) from %v", msg, c.RemoteAddr())
		return 0
	default:
		log.W("S(m)->unknow command(%v)", cmd)
		return -1
	}
}
func (r *rc_s_h) OnConn(c netw.Con) bool {
	c.SetWait(true)
	return true
}
func (r *rc_s_h) OnClose(c netw.Con) {
}

//login by remote command connection(usage 2)
func (r *rc_s_h) Login(rc *impl.RCM_Cmd) (interface{}, error) {
	var name string
	err := rc.ValidF(`
		n,O|S,L:0;
		`, &name)
	if err != nil {
		log.E("rc_s_h Login valid args error:%v", err.Error())
		return nil, err
	}
	if len(name) < 1 {
		log.E("login name is empty")
		return nil, util.Err("login name is empty")
	}
	r.L.AddC_rc(name, rc)
	log.D("S(c)->login success by name(%v)", name)
	return util.Map{
		"code": 0,
	}, nil
}

//calling by message command.
func (r *rc_s_h) CallM(target string) error {
	log.D("call1->%v", target)
	if len(target) > 0 {
		//sending command to target client.
		mc := r.L.MsgC(target)
		if mc == nil {
			return util.Err("connection not found by id(%v)", target)
		}
		_, err := mc.Writev(util.Map{
			"c": "m",
			"m": "server message",
		})
		return err
	} else {
		//sending command to all client.
		for cid, mc := range r.L.MsgCs() {
			_, err := mc.Writev(util.Map{
				"c": "m",
				"m": "server message",
			})
			if err != nil {
				log.E("sending message to %v err:%v", cid, err.Error())
			}
		}
	}
	return nil
}

//calling by remote command.
func (r *rc_s_h) CallC(target string) error {
	log.D("call2->%v", target)
	if len(target) > 0 {
		//sending command to target client.
		cc := r.L.CmdC(target)
		if cc == nil {
			return util.Err("connection not found by id(%v)", target)
		}
		var res []string
		_, err := cc.Exec("list", nil, &res)
		log.D("exec list res->%v,err:%v", res, err)
		return err
	} else {
		//sending command to all client.
		for cid, cc := range r.L.CmdCs() {
			var res []string
			_, err := cc.Exec("list", nil, &res)
			log.D("exec list to %v res->%v,err:%v", cid, res, err)
		}
	}
	return nil
}

//handler all remote command
func (r *rc_s_h) Handle(l *RC_Listener_m) {
	r.L = l
	l.AddHFunc("login", r.Login)
}

////////////////////////////////////
//////// client handler ////////////
////////////////////////////////////
type rc_c_h struct {
	R *RC_Runner_m
}

func (r *rc_c_h) OnCmd(c netw.Cmd) int {
	var args util.Map
	_, err := c.V(&args)
	if err != nil {
		log.E("rc_c_h V error:%v", err.Error())
		return -1
	}
	var cmd, msg string
	err = args.ValidF(`
		c,R|S,L:0;
		m,O|S,L:0;
		`, &cmd, &msg)
	if err != nil {
		log.E("rc_c_h valid args error:%v", err.Error())
		return -1
	}
	switch cmd {
	case "m":
		log.D("C(m)->receive message(%v) from %v", msg, c.RemoteAddr())
		return 0
	default:
		log.W("unknow command(%v)", cmd)
		return -1
	}
}
func (r *rc_c_h) OnConn(c netw.Con) bool {
	return true
}
func (r *rc_c_h) OnClose(c netw.Con) {
}

//client command
func (r *rc_c_h) List(rc *impl.RCM_Cmd) (interface{}, error) {
	log.D("C(c)->receive list command")
	return []string{"a", "b", "c"}, nil
}

//handler all client command
func (r *rc_c_h) Handle(run *RC_Runner_m) {
	r.R = run
	r.R.AddHFunc("list", r.List)
}

////////////////////////////////////
//////// testing runner ////////////
////////////////////////////////////
func TestRc(t *testing.T) {
	runtime.GOMAXPROCS(util.CPU())
	// impl.ShowLog = true
	bp := pool.NewBytePool(8, 102400)
	//
	//
	//initial server.
	sh := &rc_s_h{}
	lm := NewRC_Listener_m_j(bp, ":10801", sh)
	sh.Handle(lm)
	err := lm.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}
	//
	//
	//initial client
	crs := []*RC_Runner_m{}
	//login by message command.
	for i := 0; i < 5; i++ {
		ch := &rc_c_h{}
		cr := NewRC_Runner_m_j(bp, "127.0.0.1:10801", ch)
		ch.Handle(cr)
		cr.Start()
		_, err := cr.Writev(util.Map{
			"c": "l",
			"n": fmt.Sprintf("RC-%v", i),
		})
		_, err = cr.Writeb([]byte(util.S2Json(util.Map{
			"c": "m",
			"m": "server message",
		})))
		_, err = cr.Writev2([]byte{}, util.Map{
			"c": "m",
			"m": "server message",
		})
		if err != nil {
			t.Error(err.Error())
			return
		}
		crs = append(crs, cr)
	}
	//login by remote command.
	for i := 5; i < 10; i++ {
		ch := &rc_c_h{}
		cr := NewRC_Runner_m_j(bp, "127.0.0.1:10801", ch)
		ch.Handle(cr)
		cr.Start()
		name := fmt.Sprintf("RC-%v", i)
		res, err := cr.VExec_m("login", util.Map{
			"n": name,
		})
		if err != nil {
			t.Error(err.Error())
			return
		}
		log.D("login by name(%v)->%v", name, res.IntVal("code"))
		crs = append(crs, cr)
	}
	//
	//
	//calling target.
	for i := 0; i < 10; i++ {
		err = sh.CallM(fmt.Sprintf("RC-%v", i))
		if err != nil {
			t.Error(err.Error())
			return
		}
		err = sh.CallC(fmt.Sprintf("RC-%v", i))
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
	//calling all
	err = sh.CallM("")
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = sh.CallC("")
	if err != nil {
		t.Error(err.Error())
		return
	}
	//
	//test other
	lm.MsgC("not found")
	lm.CmdC("not found")
	lm.RCH.delc(netw.NewCon_(nil, nil, nil))
	lm.RCH.OnCmd(&netw.Cmd_{Con: netw.NewCon_(nil, nil, nil)})
	//
	//
	//close
	time.Sleep(time.Second)
	for _, cr := range crs {
		cr.Close()
		cr.Stop()
	}
	time.Sleep(time.Second)
	lm.Close()
	time.Sleep(time.Second)
}

func TestErr(t *testing.T) {
	runtime.GOMAXPROCS(util.CPU())
	bp := pool.NewBytePool(8, 102400)
	cr := NewRC_Runner_m_j(bp, "127.0.0.1:980x1", nil)
	go func() {
		fmt.Println("starting....")
		fmt.Println(cr.Writeb(nil))
		fmt.Println(cr.Writev(nil))
		fmt.Println(cr.Writev2(nil, nil))
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cr.Waitingc())
	cr.Timeout()
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cr.Waitingc())
	cr.Timeout()
	time.Sleep(100 * time.Millisecond)
	fmt.Println(cr.Waitingc())
	cr.Timeout()
	// cr.Timeout()
	// cr.Timeout()
	cr.Start()
	time.Sleep(time.Second)
}
