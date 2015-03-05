package rotation

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type stubPolicy struct {
	triggering bool
	rotated    bool
}

func (s *stubPolicy) IsTriggering(*os.File, []byte) bool {
	return s.triggering
}

func (s *stubPolicy) Rollover(*os.File) error {
	s.rotated = true
	return nil
}

func TestFile(t *testing.T) {
	testFile(t, false)
	testFile(t, true)
}

func testFile(t *testing.T, triggering bool) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)

	name := filepath.Join(dir, "test.log")
	defer os.Remove(name)

	f := NewFile(name)
	policy := &stubPolicy{triggering: triggering}
	f.SetTriggeringPolicy(policy)
	f.SetRollingPolicy(policy)

	err = f.Open(os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write([]byte("string"))
	if err != nil {
		t.Fatal(err)
	}
	if policy.rotated != triggering {
		t.Fatal("File should not rotated")
	}
	content, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if "string" != string(content) {
		t.Fatalf("unexpected content %s", content)
	}
}
