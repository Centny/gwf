package netw

import (
	"github.com/Centny/gwf/pool"
	"net"
)

//the client connection pool.
type NConPool struct {
	Addr      string //target address.
	*LConPool        //base connection pool.
}

//new client connection pool.
func NewNConPool(p *pool.BytePool, addr string, h CmdHandler) *NConPool {
	return &NConPool{
		Addr:     addr,
		LConPool: NewLConPool(p, h),
	}
}

//dail one connection.
func (n *NConPool) Dail() error {
	con, err := net.Dial("tcp", n.Addr)
	if err != nil {
		return err
	}
	n.RunC(con)
	return nil
}