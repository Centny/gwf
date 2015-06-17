package im

import (
	"fmt"
	"github.com/Centny/gwf/im/pb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/netw"
	"github.com/Centny/gwf/util"
	"strings"
	"sync"
	"time"
)

type DoImc struct {
	Srv     string
	SrvL    bool
	Tokens  []string
	Gs      []string
	Mc      int
	PushUrl string
	PushUsr string
	Res     map[string]map[string]interface{}
	//
	m_lck sync.RWMutex
}

func NewDoImc(srv string, srvl bool, tokens []string, gs []string, mc int, purl, pusr string) *DoImc {
	return &DoImc{
		Srv:     srv,
		SrvL:    srvl,
		Tokens:  tokens,
		Gs:      gs,
		Mc:      mc,
		PushUrl: purl,
		PushUsr: pusr,
		Res:     map[string]map[string]interface{}{},
	}
}
func (d *DoImc) Do() error {
	log.D("do imc by srv(%v),srvl(%v),tokens(%v),gs(%v),mc(%v),purl(%v),pusr(%v)",
		d.Srv, d.SrvL, d.Tokens, d.Gs, d.Mc, d.PushUrl, d.PushUsr)
	if len(d.Srv) < 1 {
		return util.Err("the server addres is empty")
	}
	if len(d.Tokens) < 1 {
		return util.Err("not user token")
	}
	imcs := map[string]*IMC{}
	imcs_ := []*IMC{}
	aurs := []string{}
	var imc *IMC
	var err error
	log.D("start %v connection to server", len(d.Tokens))
	for _, token := range d.Tokens {
		imc, err = d.New(token)
		if err != nil {
			return err
		}
		if _, ok := imcs[imc.IC.R]; ok {
			return util.Err("having repeat token or having two token belong to one user")
		}
		imcs[imc.IC.R] = imc
		imcs_ = append(imcs_, imc)
		aurs = append(aurs, imc.IC.R)
		d.Res[imc.IC.R] = map[string]interface{}{
			"Token": token,
		}
	}
	log.D("list group user by (%v)", d.Gs)
	gss, err := imc.GR(d.Gs)
	if err != nil {
		return err
	}
	if len(gss) != len(d.Gs) {
		fmt.Println(len(gss), len(d.Gs))
		return util.Err("having invalid group by(%v),res(%v)", d.Gs, gss)
	}
	// send the other
	log.D("sending %v message to each other", d.Mc)
	clen := len(imcs_)
	for i := 0; i < clen; i++ {
		for j := 0; j < clen; j++ {
			if i == j {
				continue
			}
			d.sms(imcs_[i], imcs_[j])
		}
	}
	log.D("sending %v message to %v group", d.Mc, len(gss))
	for gr, urs := range gss {
		d.sms_g(imcs, gr, urs)
	}
	return d.push(aurs)
	// return nil
}
func (d *DoImc) sms_g(imcs map[string]*IMC, gr string, urs []string) {
	sc := 0
	for _, ur := range urs {
		imc, ok := imcs[ur]
		if !ok {
			continue
		}
		for i := 0; i < d.Mc; i++ {
			imc.SMS(gr, 0, fmt.Sprintf("%v->%v", gr, i))
		}
		d.Res[imc.IC.R][fmt.Sprintf("S->%v", gr)] = d.Mc
		sc++
	}
	sc--
	log_d("sending %v message to R(%v)", sc*d.Mc, gr)
	if sc < 1 {
		return
	}
	for _, ur := range urs {
		if _, ok := d.Res[ur]; ok {
			d.Res[ur][fmt.Sprintf("A->%v", gr)] = sc * d.Mc
		}
	}
}
func (d *DoImc) sms(a *IMC, b *IMC) {
	for i := 0; i < d.Mc; i++ {
		a.SMS(b.IC.R, 0, fmt.Sprintf("%v->%v", b.IC.R, i))
	}
	d.Res[a.IC.R][fmt.Sprintf("S->%v", b.IC.R)] = d.Mc
	d.Res[b.IC.R][fmt.Sprintf("A->%v", a.IC.R)] = d.Mc
	log_d("sending %v messaget S(%v),R(%v)", d.Mc, a.IC.R, b.IC.R)
}
func (d *DoImc) push(aurs []string) error {
	if len(d.PushUrl) < 1 {
		return nil
	}
	log.D("doing push by url(%v),usr(%v)", d.PushUrl, d.PushUsr)
	for i := 0; i < d.Mc; i++ {
		res, err := util.HGet2(d.PushUrl, d.PushUsr, strings.Join(aurs, ","), "Push->", 0)
		if err != nil {
			return err
		}
		if res.IntVal("code") != 0 {
			return util.Err("do push to %v err:%v", d.PushUrl, res)
		}
	}
	log.D("push %v message to %v user", d.Mc, len(aurs))
	for _, ur := range aurs {
		d.Res[ur][fmt.Sprintf("A->%v", d.PushUsr)] = d.Mc
	}
	return nil
}
func (d *DoImc) OnM(i *IMC, c netw.Cmd, m *pb.ImMsg) int {
	d.m_lck.Lock()
	defer d.m_lck.Unlock()
	log_d("receive message R(%v),A(%v)", i.IC.R, m.GetA())
	v, _ := d.Res[i.IC.R][fmt.Sprintf("R->%v", m.GetA())].(int)
	d.Res[i.IC.R][fmt.Sprintf("R->%v", m.GetA())] = v + 1
	return 0
}
func (d *DoImc) New(token string) (*IMC, error) {
	log_d("New IMC by srv(%v),srvl(%v),token(%v)", d.Srv, d.SrvL, token)
	imc, err := NewIMC5(d.Srv, d.SrvL, token)
	if err != nil {
		return nil, err
	}
	imc.OnM = d.OnM
	imc.HbLog = false
	imc.Start()
	imc.LC.Wait()
	if imc.Logined() {
		imc.StartHB()
		return imc, nil
	} else {
		return nil, util.Err("login to(%v,%v) fail by token(%v)", d.Srv, d.SrvL, token)
	}
}
func (d *DoImc) Check() bool {
	for sr, res := range d.Res {
		for rk, v := range res {
			if !strings.HasPrefix(rk, "A->") {
				continue
			}
			tr := strings.TrimPrefix(rk, "A->")
			if v != res[fmt.Sprintf("R->%v", tr)] {
				log_d("checking R(%v),A(%v)->S(%v),R(%v)", sr, tr, v, res[fmt.Sprintf("R->%v", tr)])
				return false
			}
		}
	}
	return true
}
func (d *DoImc) Check2(delay, timeout int64) error {
	var used int64 = 0
	for !d.Check() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		used += delay
		if used >= timeout {
			return util.Err("timeout")
		}
	}
	return nil
}

// func (d *DoImc) Assert() error {
// 	return nil
// }
