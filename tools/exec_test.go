package tools

import (
	"fmt"
	"github.com/Centny/gwf/routing/httptest"
	"github.com/Centny/gwf/util"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	runtime.GOMAXPROCS(util.CPU())
	os.Remove("res.json")
	os.Remove("emma.xml")
	exc := NewExec("/bin/bash", "-c", "echo abc")
	exc.ShowLog = true
	exc.Run(1)
	time.Sleep(time.Second)
	exc.Wait()
	fmt.Println(exc.Res["0"])
	exc = NewExec("sdfs/sdf", "...")
	exc.Run(1)
	exc.Execing()
	exc.Wait()
	//
	exk := NewExeK(1, 10, 25, "/bin/bash", "-c", "xx")
	exk.MT = 100000000
	exk.Start()
	exk.Wait()
	if exk.DoneSize() != 25 {
		fmt.Println("size:", exk.DoneSize())
		t.Error("not right")
		return
	}
	//
	exk = NewExeK(1, 10, 25, "/bin/bash", "-c")
	exk.CmdF = func(exe *Exec, exk *ExeK, idx string) *exec.Cmd {
		args := append(exe.Args, fmt.Sprintf("echo v-%v", idx))
		return exec.Command(exe.Bin, args...)
	}
	exk.Start()
	exk.Wait()
	if exk.DoneSize() != 25 {
		fmt.Println("size:", exk.DoneSize())
		t.Error("not right")
		return
	}
	exk.SaveP("res.json", "emma.xml")
	exk.Save(os.Stdout)
	//
	//
	ts := httptest.NewMuxServer()
	ts.Mux.HFunc("^/list", exk.List)
	ts.Mux.HFunc("^/logs", exk.Logs)
	fmt.Println(ts.G("/list"))
	fmt.Println(ts.G("/logs?id=0&key=o_out"))
	fmt.Println(ts.G("/logs"))
	fmt.Println(ts.G("/logs?id=xx&key=o_out"))
	//
	exk = NewExeK(1, 10, 25, "/bin/bash", "-c")
	exk.Start()
	exk.Wait()
	//
}
