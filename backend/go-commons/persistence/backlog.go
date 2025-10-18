package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

type EventRecord struct {
	Timestamp string          `json:"timestamp"`
	Stream    string          `json:"stream"`
	Subject   string          `json:"subject"`
	MsgID     string          `json:"msgID"`
	Data      json.RawMessage `json:"data"`
}

type Backlog struct {
	dir       string
	maxBytes  int64
	outFile   *os.File
	size      int64
	seq       uint64
	pid       int
	writes    int
	syncEvery int
}

func NewBacklog(dir string, maxBytes int64, syncEvery int) *Backlog {
	if syncEvery <= 0 {
		syncEvery = 1
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logrus.Panic(fmt.Sprintf("cannot create backlog dir %s: %v", dir, err))
	}

	return &Backlog{
		dir:       dir,
		maxBytes:  maxBytes,
		pid:       os.Getpid(),
		syncEvery: syncEvery,
	}
}

func (b *Backlog) filename() string {
	ts := time.Now().UTC().Format("20060102-150405.000000000")
	return filepath.Join(b.dir, fmt.Sprintf("backlog-%s-%d-%06d.jsonl", ts, b.pid, b.seq))
}

// func (b *Backlog) rotateIfNeeded() error {
// 	if b.outFile == nil || b.size >= b.maxBytes {
// 		if b.outFile != nil {
// 			_ = b.outFile.Sync()
// 			_ = b.outFile.Close()
// 			b.outFile = nil
// 		}
// 		name := b.filename()
// 		f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
// 		if err != nil {
// 			return err
// 		}
// 		info, err := f.Stat()
// 		if err != nil {
// 			_ = f.Close()
// 			return err
// 		}
// 		b.outFile = f
// 		b.size = info.Size()
// 		b.seq++
// 		b.writes = 0
// 	}
// 	return nil
// }

func (b *Backlog) rotateIfNeeded() {
	// si el archivo ya existe y se alcanzó el tamaño máximo → cerrarlo
	if b.outFile != nil && b.size >= b.maxBytes {
		_ = b.outFile.Sync()
		_ = b.outFile.Close()
		b.outFile = nil
		b.size = 0
		b.seq++
		b.writes = 0
	}
}

func (b *Backlog) RotateNow() error {
	if b.outFile != nil {
		_ = b.outFile.Sync()
		_ = b.outFile.Close()
		b.outFile = nil
	}
	b.size = 0
	return nil
}

func (b *Backlog) pendingFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(b.dir, "backlog-*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (b *Backlog) HasPending() bool {
	files, err := b.pendingFiles()
	return err == nil && len(files) > 0
}

func (b *Backlog) Write(ev EventRecord) error {
	b.rotateIfNeeded()

	// ⚠️ Lazy open: abrir archivo solo si es necesario
	if b.outFile == nil {
		name := b.filename()
		f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		b.outFile = f
		b.size = 0
		b.writes = 0
	}

	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	n, err := b.outFile.Write(append(line, '\n'))
	if err != nil {
		return err
	}
	b.size += int64(n)
	b.writes++
	if b.writes%b.syncEvery == 0 {
		_ = b.outFile.Sync()
	}
	return nil
}

func (b *Backlog) Replay(callback func(ev EventRecord) error) error {
	files, err := b.pendingFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		tmpFile := file + ".tmp"

		in, err := os.Open(file)
		if err != nil {
			continue
		}
		out, err := os.Create(tmpFile)
		if err != nil {
			_ = in.Close()
			return err
		}

		scanner := bufio.NewScanner(in)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			var ev EventRecord
			if json.Unmarshal(scanner.Bytes(), &ev) == nil {
				if err := callback(ev); err != nil {
					// fallo → lo guardamos para reintentar
					out.Write(scanner.Bytes())
					out.Write([]byte("\n"))
				}
			}
		}

		_ = in.Close()
		_ = out.Close()

		// verificamos si quedaron pendientes
		info, _ := os.Stat(tmpFile)
		if info != nil && info.Size() == 0 {
			// todo publicado bien → borramos backlog y tmp
			_ = os.Remove(tmpFile)
			_ = os.Remove(file)
		} else {
			// quedaron pendientes → reemplazamos backlog con el tmp
			_ = os.Rename(tmpFile, file)
		}
	}

	return nil
}

func (b *Backlog) ReplayBatched(
	batchSize int,
	publish func(batch []EventRecord) []bool, // devuelve éxito por item
) error {
	files, err := b.pendingFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		tmpFile := file + ".tmp"

		in, err := os.Open(file)
		if err != nil {
			continue
		}
		out, err := os.Create(tmpFile)
		if err != nil {
			_ = in.Close()
			return err
		}

		sc := bufio.NewScanner(in)
		buf := make([]byte, 0, 1024*1024)
		sc.Buffer(buf, 10*1024*1024)

		type rawLine struct {
			bytes []byte
			ev    EventRecord
		}
		var batch []rawLine

		flush := func() error {
			if len(batch) == 0 {
				return nil
			}
			records := make([]EventRecord, len(batch))
			for i := range batch {
				records[i] = batch[i].ev
			}
			ok := publish(records) // len(ok) == len(batch)

			for i, success := range ok {
				if !success {
					out.Write(batch[i].bytes)
					out.Write([]byte("\n"))
				}
			}
			batch = batch[:0]
			return nil
		}

		for sc.Scan() {
			line := append([]byte(nil), sc.Bytes()...) // copy
			var ev EventRecord
			if json.Unmarshal(line, &ev) == nil {
				batch = append(batch, rawLine{bytes: line, ev: ev})
				if len(batch) >= batchSize {
					if err := flush(); err != nil {
						_ = in.Close()
						_ = out.Close()
						return err
					}
				}
			}
		}
		_ = flush()

		_ = in.Close()
		_ = out.Close()

		// ¿quedó algo pendiente?
		info, _ := os.Stat(tmpFile)
		if info != nil && info.Size() == 0 {
			_ = os.Remove(tmpFile)
			_ = os.Remove(file)
		} else {
			_ = os.Rename(tmpFile, file)
		}
	}
	return nil
}

func (b *Backlog) Close() error {
	if b.outFile != nil {
		_ = b.outFile.Sync()
		err := b.outFile.Close()
		b.outFile = nil
		return err
	}
	return nil
}
