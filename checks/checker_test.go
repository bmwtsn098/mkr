package checks

import (
	"bytes"
	"testing"
)

type testChecker struct {
	*result
}

func (te *testChecker) check() *result {
	return te.result
}

func TestRunChecks(t *testing.T) {
	te := &testChecker{&result{
		Name:     "hoge",
		Cmd:      []string{"perl", "-E", "say 'Hello'"},
		Status:   "OK",
		Stdout:   "Hello",
		ExitCode: 0,
	}}
	buf := &bytes.Buffer{}
	if err := runChecks([]checker{te}, buf); err != nil {
		t.Fatal(err)
	}

	expect := `TAP version 13
1..1
ok 1 - hoge
  ---
  command: [perl, -E, say 'Hello']
  status: OK
  stdout: Hello
  ...
`
	got := buf.String()
	if got != expect {
		t.Errorf("something went wrong\ngot:\n%s\nexpect:\n%s", got, expect)
	}
}
