package file

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultFilePattern(t *testing.T) {
	name := "/var/log/gol.log"
	target := defaultFilePattern(name)
	if "/var/log/gol-%s.log.gz" != target {
		t.Fatalf(target)
	}
}

func TestFileExists(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	exist := fileExists(f.Name())
	f.Close()
	os.Remove(f.Name())

	if !exist {
		t.Fatalf("file does not exist %#v", f.Name())
	}

	exist = fileExists(f.Name())
	if exist {
		t.Fatalf("file exists %#v", f.Name())
	}
}

func stubCurrentTime() time.Time {
	return time.Date(2015, 4, 3, 0, 0, 0, 0, time.Local)
}

func TestTimeRollingPolicy(t *testing.T) {
	// Create a directory for logging
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(dir)
	}()
	// Log file
	name := filepath.Join(dir, "test.log")
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(name)
	}()
	_, err = f.Write([]byte("line 1\nline 2\n"))
	if err != nil {
		t.Fatal(err)
	}
	// Rolling policy
	p := NewTimeRollingPolicy()
	p.TimeKeeper = timeKeeperFunc(stubCurrentTime)
	p.FilePattern = filepath.Join(dir, "archive-%s.log")
	err = p.Rollover(f)
	if err != nil {
		t.Fatal(err)
	}
	// Archive must have log content
	archiveName := filepath.Join(dir, "archive-2015-04-03.log")
	defer func() {
		os.Remove(archiveName)
	}()

	content, err := ioutil.ReadFile(archiveName)
	if err != nil {
		t.Fatal(err)
	}
	if "line 1\nline 2\n" != string(content) {
		t.Fatalf("unexpected content: %#v", string(content))
	}
	// Log file must be empty
	content, err = ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != 0 {
		t.Fatalf("unexpected content: %#v", string(content))
	}
}

func TestTimeRollingPolicyGzip(t *testing.T) {
	// Log file
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	_, err = f.Write([]byte("line 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	// Policy
	p := NewTimeRollingPolicy()
	p.TimeKeeper = timeKeeperFunc(stubCurrentTime)
	err = p.Rollover(f)
	if err != nil {
		t.Fatal(err)
	}
	// TempFile does not create file with an extension
	archiveFile, err := os.Open(f.Name() + "-2015-04-03.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		archiveFile.Close()
		os.Remove(archiveFile.Name())
	}()
	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		t.Fatal(err)
	}
	if "line 1\n" != string(content) {
		t.Fatalf("unexpected gzip content: %#v", string(content))
	}
}

func TestTimeRollingPolicyHistory(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(dir)
	}()
	name := filepath.Join(dir, "test.log")
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	name1 := filepath.Join(dir, "test-2015-04-02.log.gz")
	if err = ioutil.WriteFile(name1, []byte("20150402"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	name2 := filepath.Join(dir, "test-2015-04-01.log.gz")
	if err = ioutil.WriteFile(name2, []byte("20150401"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	name3 := filepath.Join(dir, "test-2015-03-31.log.gz")
	if err = ioutil.WriteFile(name3, []byte("20150331"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	p := NewTimeRollingPolicy()
	p.TimeKeeper = timeKeeperFunc(stubCurrentTime)
	p.FilePattern = filepath.Join(dir, "test-%s.log.gz")
	p.FileCount = 4
	if err = p.Rollover(f); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(name3)
	if err != nil {
		t.Fatal(err)
	}
	if "20150331" != string(content) {
		t.Fatalf("unexpected content: %#v", string(content))
	}
	p.FileCount = 3
	if err = p.Rollover(f); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name3); err == nil {
		t.Fatal("%#v should be removed", name3)
	}
	p.FileCount = 2
	if err = p.Rollover(f); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name2); err == nil {
		t.Fatal("%#v should be removed", name2)
	}
	p.FileCount = 1
	if err = p.Rollover(f); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name1); err == nil {
		t.Fatal("%#v should should be removed", name1)
	}
}
