package checker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mackerelio/mackerel-agent/config"
	cli "gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

var Command = cli.Command{
	Name:  "run-checks",
	Usage: "run check commands in mackerel-agent.conf",
	Description: `
    Execute command of check plugins in mackerel-agent.conf all at once.
    It is used for checking setting and operation of the check plugins.
`,
	Action: doRunChecks,
}

func doRunChecks(c *cli.Context) error {
	confFile := c.GlobalString("conf")
	conf, err := config.LoadConfig(confFile)
	if err != nil {
		return err
	}
	checkers := make([]checker, len(conf.CheckPlugins))
	i := 0
	for name, p := range conf.CheckPlugins {
		checkers[i] = &checkPluginChecker{
			name: name,
			cp:   p,
		}
		i++
	}
	return runChecks(checkers, os.Stdout)
}

type result struct {
	Name     string `yaml:"-"`
	Memo     string `yaml:"memo,omitempty"`
	Cmd      string `yaml:"command"`
	Stdout   string `yaml:"stdout,omitempty"`
	Stderr   string `yaml:"stderr,omitempty"`
	ExitCode int    `yaml:"exitCode,omitempty"`
	ErrMsg   string `yaml:"error,omitempty"`
}

func (re *result) tapFormat(num int) string {
	okOrNot := "ok"
	if re.ExitCode != 0 || re.ErrMsg != "" {
		okOrNot = "not ok"
	}
	b, _ := yaml.Marshal(re)
	// indent
	yamlStr := "  " + strings.Replace(strings.TrimSpace(string(b)), "\n", "\n  ", -1)
	return fmt.Sprintf("%s %d - %s\n  ---\n%s\n  ...",
		okOrNot, num, re.Name, yamlStr)
}

type checkPluginChecker struct {
	name string
	cp   *config.CheckPlugin
}

func (cpc *checkPluginChecker) check() *result {
	p := cpc.cp
	stdout, stderr, exitCode, err := p.Command.Run()
	cmdStr := p.Command.Cmd
	if cmdStr == "" {
		b, _ := json.Marshal(p.Command.Args)
		cmdStr = string(b)
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &result{
		Name:     cpc.name,
		Memo:     p.Memo,
		Cmd:      cmdStr,
		ExitCode: exitCode,
		Stdout:   strings.TrimSpace(stdout),
		Stderr:   strings.TrimSpace(stderr),
		ErrMsg:   errMsg,
	}
}

type checker interface {
	check() *result
}

func runChecks(checkers []checker, w io.Writer) error {
	ch := make(chan *result)
	total := len(checkers)
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(total)
		for _, c := range checkers {
			go func(c checker) {
				defer wg.Done()
				ch <- c.check()
			}(c)
		}
		wg.Wait()
		close(ch)
	}()
	fmt.Fprintln(w, "TAP version 13")
	fmt.Fprintf(w, "1..%d\n", total)
	testNum := 1
	for re := range ch {
		fmt.Fprintln(w, re.tapFormat(testNum))
		testNum++
	}
	return nil
}
