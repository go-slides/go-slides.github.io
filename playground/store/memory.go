package store

import (
	"errors"
	"time"
)

// Memory is a Backend held in memory, shared by the stores given the same
// one (the tabs of a browser, in tests): it answers at once. Quota bounds
// the bytes it holds (0: none); Blocked refuses everything, as a browser
// that blocks storage does.
type Memory struct {
	Quota   int
	Blocked bool
	saved   map[string]*Project
}

// NewMemory returns an empty Memory.
func NewMemory() *Memory { return &Memory{saved: map[string]*Project{}} }

var errBlocked = errors.New("SecurityError")

// ErrQuota is what a full Memory answers, as IndexedDB's QuotaExceededError.
var ErrQuota = errors.New("QuotaExceededError")

func (m *Memory) copyOf(p *Project, name string) *Project {
	files := map[string]File{}
	for k, f := range p.files {
		files[k] = File{Data: append([]byte(nil), f.Data...), Modified: f.Modified}
	}
	return Restore(name, p.Saved, p.Tab, p.UI, files)
}

// Load implements Backend.
func (m *Memory) Load(name string, done func(*Project, error)) {
	if m.Blocked {
		done(nil, errBlocked)
		return
	}
	p, ok := m.saved[name]
	if !ok {
		done(nil, nil)
		return
	}
	done(m.copyOf(p, name), nil)
}

func (m *Memory) size(except string) int {
	n := 0
	for k, p := range m.saved {
		if k != except {
			n += p.Size()
		}
	}
	return n
}

// Save implements Backend.
func (m *Memory) Save(p *Project, written, removed []string, done func(error)) {
	if m.Blocked {
		done(errBlocked)
		return
	}
	cur, ok := m.saved[p.Name]
	if !ok {
		cur = NewProject(p.Name)
	} else {
		cur = m.copyOf(cur, p.Name)
	}
	for _, w := range written {
		if f, ok := p.files[w]; ok {
			cur.files[w] = File{Data: append([]byte(nil), f.Data...), Modified: f.Modified}
		}
	}
	for _, r := range removed {
		delete(cur.files, r)
	}
	if m.Quota > 0 && m.size(p.Name)+cur.Size() > m.Quota {
		done(ErrQuota) // the transaction aborts: nothing is written
		return
	}
	cur.Saved, cur.Tab = p.Saved, p.Tab
	cur.UI = map[string]string{}
	for k, v := range p.UI {
		cur.UI[k] = v
	}
	cur.clean()
	m.saved[p.Name] = cur
	done(nil)
}

// Copy implements Backend.
func (m *Memory) Copy(from, to string, done func(error)) {
	if m.Blocked {
		done(errBlocked)
		return
	}
	p, ok := m.saved[from]
	if !ok {
		delete(m.saved, to)
		done(nil)
		return
	}
	if m.Quota > 0 && m.size(to)+p.Size() > m.Quota {
		done(ErrQuota)
		return
	}
	m.saved[to] = m.copyOf(p, to)
	done(nil)
}

// Delete implements Backend.
func (m *Memory) Delete(name string, done func(error)) {
	if m.Blocked {
		done(errBlocked)
		return
	}
	delete(m.saved, name)
	done(nil)
}

// Saved reports when a project was saved, for tests.
func (m *Memory) Saved(name string) (time.Time, bool) {
	p, ok := m.saved[name]
	if !ok {
		return time.Time{}, false
	}
	return p.Saved, true
}
