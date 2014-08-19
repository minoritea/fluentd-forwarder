package fluentd_forwarder

import (
	"testing"
	"time"
	"math/rand"
	logging "github.com/op/go-logging"
	"os"
	"fmt"
	"io/ioutil"
)

type DummyWorker struct { v int }

func (*DummyWorker) String() string { return "" }
func (*DummyWorker) Start() {}
func (*DummyWorker) Stop() {}
func (*DummyWorker) WaitForShutdown() {}

type DummyChunkListener struct {
	t *testing.T
	chunks []JournalChunk
}

func (*DummyChunkListener) NewChunkCreated(chunk JournalChunk) error {
	return nil
}

func (listener *DummyChunkListener) ChunkFlushed(chunk JournalChunk) error {
	listener.chunks = append(listener.chunks, chunk)
	listener.t.Logf("flush %d", len(listener.chunks))
	return nil
}

func Test_GetJournalGroup(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		0,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup1, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journalGroup2, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.Fail() }
	if journalGroup1 != journalGroup2 { t.Fail() }
	anotherDummyWorker := &DummyWorker {}
	if dummyWorker == anotherDummyWorker { t.Log("WTF?"); t.Fail() }
	_, err = factory.GetJournalGroup(tempDir + "/test", anotherDummyWorker)
	if err == nil { t.Fail() }
}

func Test_Journal_GetJournal(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		0,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journal1 := journalGroup.GetJournal("key")
	if journal1 == nil { t.FailNow() }
	journal2 := journalGroup.GetJournal("key")
	if journal2 == nil { t.Fail() }
	if journal1 != journal2 { t.Fail() }
}

func Test_Journal_EmitVeryFirst(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		10,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journal := journalGroup.GetFileJournal("key")
	err = journal.Write([]byte("test"))
	if err != nil { t.FailNow() }
	if journal.chunks.count != 1 { t.Fail() }
	if journal.chunks.first.Size != 4 { t.Fail() }
}

func Test_Journal_EmitTwice(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		10,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journal := journalGroup.GetFileJournal("key")
	err = journal.Write([]byte("test1"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test2"))
	if err != nil { t.FailNow() }
	if journal.chunks.count != 1 { t.Fail() }
	if journal.chunks.first.Size != 10 { t.Fail() }
}

func Test_Journal_EmitRotating(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		8,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journal := journalGroup.GetFileJournal("key")
	err = journal.Write([]byte("test1"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test2"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test3"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test4"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test5"))
	if err != nil { t.FailNow() }
	t.Logf("journal.chunks.count=%d", journal.chunks.count)
	t.Logf("journal.chunks.first.Size=%d", journal.chunks.first.Size)
	if journal.chunks.count != 5 { t.Fail() }
	if journal.chunks.first.Size != 5 { t.Fail() }
}

func shuffle(x []string) {
	rng := rand.New(rand.NewSource(0))
	for i := 0; i < len(x); i += 1 {
		j := rng.Intn(i + 1)
		x[i], x[j] = x[j], x[i]
	}
}

func Test_Journal_Scanning_Ok(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tm := time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 1; i < 100; i++ {
		tempDir, err := ioutil.TempDir("", "journal")
		if err != nil { t.FailNow() }
		prefix := tempDir + "/test"
		suffix := ".log"
		makePaths := func (n int, key string) []string {
			paths := make([]string, n)
			for i := 0; i < len(paths); i += 1 {
				type_ := JournalFileType('q')
				if i == 0 {
					type_ = JournalFileType('b')
				}
				path := prefix + "." + BuildJournalPath(key, type_, tm.Add(time.Duration(-i * 1e9)), 0).VariablePortion + suffix
				paths[i] = path
			}
			return paths
		}

		paths := makePaths(i, "key")
		shuffledPaths := make([]string, len(paths))
		copy(shuffledPaths, paths)
		shuffle(shuffledPaths)
		for j, path := range shuffledPaths {
			file, err := os.Create(path)
			if err != nil {
				t.FailNow()
			}
			_, err = file.Write([]byte(fmt.Sprintf("%08d", j)))
			if err != nil {
				t.FailNow()
			}
			file.Close()
		}
		factory := NewFileJournalGroupFactory(
			logger,
			rand.NewSource(0),
			func () time.Time { return tm },
			suffix,
			os.FileMode(0644),
			8,
		)
		dummyWorker := &DummyWorker {}
		journalGroup, err := factory.GetJournalGroup(prefix, dummyWorker)
		if err != nil { t.FailNow() }
		journal := journalGroup.GetFileJournal("key")
		t.Logf("journal.chunks.count=%d", journal.chunks.count)
		t.Logf("journal.chunks.first.Size=%d", journal.chunks.first.Size)
		if journal.chunks.count != i { t.Fail() }
		j := 0
		for chunk := journal.chunks.first; chunk != nil; chunk = chunk.head.next {
			if chunk.Path != paths[j] { t.Fail() }
			j += 1
		}
		journal.Purge()
		if journal.chunks.count != 1 { t.Fail() }
	}
}

func Test_Journal_Scanning_MultipleHead(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	tm := time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)
	prefix := tempDir + "/test"
	suffix := ".log"
	createFile := func (key string, type_ JournalFileType, o int) (string, error) {
		path := prefix + "." + BuildJournalPath(key, type_, tm.Add(time.Duration(-o * 1e9)), 0).VariablePortion + suffix
		file, err := os.Create(path)
		if err != nil {
			return "", err
		}
		_, err = file.Write([]byte(fmt.Sprintf("%08d", o)))
		if err != nil {
			return "", err
		}
		file.Close()
		t.Log(path)
		return path, nil
	}

	paths := make([]string, 4)
	{
		path, err := createFile("key", JournalFileType('b'), 0)
		if err != nil { t.FailNow() }
		paths[0] = path
	}
	{
		path, err := createFile("key", JournalFileType('b'), 1)
		if err != nil { t.FailNow() }
		paths[1] = path
	}
	{
		path, err := createFile("key", JournalFileType('q'), 2)
		if err != nil { t.FailNow() }
		paths[2] = path
	}
	{
		path, err := createFile("key", JournalFileType('q'), 3)
		if err != nil { t.FailNow() }
		paths[3] = path
	}

	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return tm },
		suffix,
		os.FileMode(0644),
		8,
	)
	dummyWorker := &DummyWorker {}
	_, err = factory.GetJournalGroup(prefix, dummyWorker)
	if err == nil { t.FailNow() }
	t.Log(err.Error())
	if err.Error() != "multiple chunk heads found" { t.Fail() }
}

func Test_Journal_FlushListener(t *testing.T) {
	logging.InitForTesting(logging.NOTICE)
	logger := logging.MustGetLogger("journal")
	tempDir, err := ioutil.TempDir("", "journal")
	if err != nil { t.FailNow() }
	factory := NewFileJournalGroupFactory(
		logger,
		rand.NewSource(0),
		func () time.Time { return time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) },
		".log",
		os.FileMode(0644),
		8,
	)
	dummyWorker := &DummyWorker {}
	t.Log(tempDir + "/test")
	journalGroup, err := factory.GetJournalGroup(tempDir + "/test", dummyWorker)
	if err != nil { t.FailNow() }
	journal := journalGroup.GetFileJournal("key")
	listener := &DummyChunkListener {
		t: t,
		chunks: make([]JournalChunk, 0, 5),
	}
	journal.AddFlushListener(listener)
	journal.AddFlushListener(listener)
	err = journal.Write([]byte("test1"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test2"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test3"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test4"))
	if err != nil { t.FailNow() }
	err = journal.Write([]byte("test5"))
	if err != nil { t.FailNow() }
	t.Logf("journal.chunks.count=%d", journal.chunks.count)
	t.Logf("journal.chunks.first.Size=%d", journal.chunks.first.Size)
	if journal.chunks.count != 5 { t.Fail() }
	if journal.chunks.first.Size != 5 { t.Fail() }
	if len(listener.chunks) != 4 { t.Fail() }
	readAll := func (chunk JournalChunk) string {
		reader, err := chunk.GetReader()
		if err != nil { t.FailNow() }
		bytes, err := ioutil.ReadAll(reader)
		if err != nil { t.FailNow() }
		return string(bytes)
	}
	if readAll(listener.chunks[0]) != "test1" { t.Fail() }
	if readAll(listener.chunks[1]) != "test2" { t.Fail() }
	if readAll(listener.chunks[2]) != "test3" { t.Fail() }
	if readAll(listener.chunks[3]) != "test4" { t.Fail() }
	journal.Purge()
	if journal.chunks.count != 5 { t.Fail() }
	for _, chunk := range listener.chunks {
		chunk.Dispose()
	}
	t.Logf("journal.chunks.count=%d", journal.chunks.count)
	t.Logf("journal.chunks.first.Size=%ld", journal.chunks.first.Size)
	journal.Purge()
	if journal.chunks.count != 1 { t.Fail() }
	if journal.chunks.first.Type != JournalFileType('b') { t.Fail() }
}

