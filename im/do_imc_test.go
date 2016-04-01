package im

import (
	"fmt"
	"github.com/Centny/gwf/netw"
	"github.com/Centny/gwf/pool"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/routing/httptest"
	"github.com/Centny/gwf/tutil"
	"net/http"
	"runtime"
	"testing"
	"time"
)

func TestDoImc(t *testing.T) {
	run_do_imc_(t, 5, 5)
}

func run_do_imc_(t *testing.T, total, tc int) {
	ShowLog = true
	netw.ShowLog_C = true
	netw.ShowLog = true
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	db := NewMemDbH()
	p := pool.NewBytePool(8, 102400)
	psrv := NewPushSrv(p, ":5498", "Push", netw.NewDoNotH(), db)
	psrv.TickLog = false
	err := psrv.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}
	ts := httptest.NewServer(func(hs *routing.HTTPSession) routing.HResult {
		var r, c, s string
		var t int64
		err := hs.ValidCheckVal(`
		s,R|S,L:0;
		r,R|S,L:0;
		c,R|S,L:0;
		t,R|I,O:0~1~101;
		`, &s, &r, &c, &t)
		if err != nil {
			return hs.MsgResErr2(1, "arg-err", err)
		}
		_, err = psrv.PushN(s, r, c, uint32(t))
		if err == nil {
			return hs.MsgRes("OK")
		} else {
			return hs.MsgResErr2(1, "srv-err", err)
		}
	})
	l := NewListner3(db, fmt.Sprintf("S-vv-%v", 0), p, 9780, 1000000)
	l.WsAddr = fmt.Sprintf(":%v", 9770)
	l.PushSrvAddr = "127.0.0.1:5498"
	l.PushSrvTickLog = false
	rc := make(chan int)
	go func() {
		rc <- 1
		hs := &http.Server{
			Handler: l.WIM_L.WsS(),
			Addr:    l.WsAddr,
		}
		hs.ListenAndServe()
	}()
	err = l.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}
	<-rc
	time.Sleep(time.Second)
	pool.BP = pool.NewBytePool(8, 10240000)
	//
	purl := ts.URL + "?s=%v&r=%v&c=%v&t=%v"
	tutil.DoPerfV(total, tc, "", func(i int) {
		run_do_imc_c(i, db, purl, t)
	})
	if len(db.Ms) > 0 || len(db.Cons) > 0 || len(db.Usr) > 0 ||
		len(db.Grp) > 0 || len(db.Tokens) > 0 || len(db.U2M) > 0 {
		t.Error("error")
		return
	}
	fmt.Println("all done...")
}
func run_do_imc_c(i int, db *MemDbH, purl string, t *testing.T) {
	ga := fmt.Sprintf("G-%v", i)
	ua, ub, uc, ud := fmt.Sprintf("U-%v-%v", i, 1), fmt.Sprintf("U-%v-%v", i, 2),
		fmt.Sprintf("U-%v-%v", i, 3), fmt.Sprintf("U-%v-%v", i, 4)
	ta, tb, tc := fmt.Sprintf("T-%v-%v", i, 1), fmt.Sprintf("T-%v-%v", i, 2), fmt.Sprintf("T-%v-%v", i, 3)
	db.AddGrp(ga, []string{
		ua, ub, uc,
		"U-abc",
	})
	db.AddTokens(map[string]string{
		ta: ua,
		tb: ub,
		tc: uc,
	})
	di := NewDoImc(pool.BP, ":9780", false, []string{ta, tb, tc}, []string{ga}, 8, purl, ud)

	err := di.Do()
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = di.Check2(1000, 10000)
	if err != nil {
		t.Error(err.Error())
		fmt.Println(di.Res)
		return
	}
	di.Release()
	db.DelTokens([]string{ta, tb, tc})
	db.DelGrp(ga)
	db.ClearMsg([]string{ua, ub, uc, ud})
}
