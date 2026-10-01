package tool

import (
	"context"
	"net"
	"os/exec"
	"time"

	"github.com/tatsushid/go-fastping"
	"github.com/zendev-sh/goai"
)

// 给树莓派用的，通过GPIO 18实现台式机开机和关机
func ControlPc() goai.Tool {
	return goai.NewTool(
		"ControlPc",
		"控制电脑开机和关机",
		func(ctx context.Context, input struct {
			Option int `json:"option" jsonschema:"description=0是关机，1是开机"`
		}) (string, error) {
			p := fastping.NewPinger()
			ra, err := net.ResolveIPAddr("ip4:icmp", "pc")
			if err != nil {
				return "", err
			}

			p.AddIPAddr(ra)
			var pingOK bool
			p.OnRecv = func(i *net.IPAddr, d time.Duration) {
				pingOK = true
			}

			if err := p.Run(); err != nil {
				return "", err
			}
			if (pingOK && input.Option == 1) || (!pingOK && input.Option == 0) {
				return "电脑已经是指定状态，不需要这次操作", nil
			}

			if err = exec.Command("sh", "-c", "sudo gpioset -c 0 -t 800ms,0 18=1").Run(); err != nil {
				return "", err
			}

			return "操作已完成,请稍等", nil
		})
}
