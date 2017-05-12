package state

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/rollerderby/go/json"
	"github.com/rollerderby/go/logger"
)

var (
	log  = *logger.New("state")
	Root = &root{values: make(map[string]*rootValue), basePath: "."}
)

type rootValue struct {
	backingFile string
	value       Value
}

type root struct {
	values   map[string]*rootValue
	revision uint64
	basePath string

	mu       sync.Mutex // Lock the state so only one "changer" can update the state at a time
	isLocked bool       // Not foolproof, but helps to detect someone updating the state without holding the lock
	changed  bool       // Was the state changed between locks?
	isReady  bool       // Is the system up and ready to go?
}

func (r *root) Lock() {
	r.mu.Lock()
	r.isLocked = true
}

func (r *root) Unlock() {
	if r.changed {
		r.revision++
		r.changed = false
	}
	r.isLocked = false
	r.mu.Unlock()
}

func (r *root) Get(key string) Value {
	val := r.values[key]
	if val == nil {
		return nil
	}
	return val.value
}

func (r *root) IsReady() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isReady
}

func (r *root) SetIsReady(isReady bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.isReady = isReady
}

func (r *root) changedValue(value Value) {
	if value.Parent() == nil {
		return
	}

	if !r.isLocked {
		log.Critf("Changing the state without holding a lock!  Value path: %v", value.Path())
		for i := 1; i < 5; i++ {
			_, file, no, ok := runtime.Caller(i)
			if ok {
				log.Critf("  Called from %s#%d\n", file, no)
			} else {
				break
			}
		}
	}

	value.SetRevision(r.revision)
	r.changed = true

	for value.Parent() != r {
		value.SetSaveNeeded(true)
		value = value.Parent()
	}
}

func (r *root) WriteGroups() []string         { return nil }
func (r *root) AddWriteGroup(group ...string) {}
func (r *root) ReadGroups() []string          { return nil }
func (r *root) AddReadGroup(group ...string)  {}
func (r *root) SaveNeeded() bool              { return false }
func (r *root) SetSaveNeeded(skip bool)       {}
func (r *root) SkipSave() bool                { return false }
func (r *root) SetSkipSave(skip bool)         {}
func (r *root) Revision() uint64              { return r.revision }
func (r *root) SetRevision(rev uint64)        {}

func (r *root) Parent() Value { return nil }
func (r *root) Path() string  { return "" }

func (r *root) SetParentAndPath(parent Value, path string) {}

func (r *root) JSON(skipSave bool) json.Value {
	j := make(json.Object)
	for key, value := range r.values {
		j[key] = value.value.JSON(skipSave)
	}

	return j
}

func (r *root) SetJSON(j json.Value) error {
	return errNotImplemented
}

func (r *root) String() string {
	return "ROOT"
}

func (r *root) Add(name, backingFile string, value Value) error {
	if _, ok := r.values[name]; ok {
		return errExistingKey(name)
	}

	r.values[name] = &rootValue{
		backingFile: backingFile,
		value:       value,
	}

	value.SetParentAndPath(r, name)

	return nil
}

func (r *root) LoadSavedConfigs() error {
	// Take lock so we can load config files in to the active state
	r.Lock()
	defer r.Unlock()

	var errors []error

	loadJSON := func(filename string) (json.Value, error) {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		return json.Decode(data)
	}

	for valueName, value := range r.values {
		if value.backingFile == "" {
			continue
		}
		if hash, ok := value.value.(*Hash); ok {
			dir := path.Join(r.basePath, "config", value.backingFile)

			files, err := ioutil.ReadDir(dir)
			if err != nil {
				errors = append(errors, err)
				continue
			}

			for _, fi := range files {
				key := fi.Name()
				key = key[:len(key)-len(path.Ext(fi.Name()))]

				json, err := loadJSON(path.Join(dir, fi.Name()))
				if err != nil {
					errors = append(errors, err)
					continue
				}

				if val := hash.Get(key); val == nil {
					log.Infof("Loading %q into hash %q as key %q", fi.Name(), valueName, key)
					if newValue, err := hash.NewElement(key, json); err != nil {
						errors = append(errors, err)
					} else {
						newValue.SetSaveNeeded(false)
					}
				}
			}
		} else {
			filename := path.Join(r.basePath, "config", value.backingFile+".json")
			log.Infof("Looking for config %q for %q", filename, valueName)

			json, err := loadJSON(filename)
			if err != nil {
				errors = append(errors, err)
				continue
			}

			log.Infof("Loading %q into %q", filename, valueName)
			if err := value.value.SetJSON(json); err != nil {
				errors = append(errors, err)
				continue
			}

			value.value.SetSaveNeeded(false)
		}
	}

	if len(errors) > 0 {
		log.Error("Errors while loading configs")
		for _, err := range errors {
			log.Errorf("   %v", err)
		}
	}
	return nil
}

// Runs in a goroutine
func (r *root) SaveLoop() {
	save := func(filename string, json json.Value) {
		dir := path.Dir(filename)
		if err := os.MkdirAll(dir, 0775); err != nil {
			log.Errorf("Cannot create director %q: %v", dir, err)
		} else if err := ioutil.WriteFile(filename, []byte(json.JSON(true)), 0664); err != nil {
			log.Errorf("Cannot save config file %q: %v", filename, err)
		} else {
			log.Infof("Saved config file %q", filename)
		}
	}
	saveAllNeeded := func() {
		// Take lock so no data is changed while we save
		r.Lock()
		defer r.Unlock()
		for _, value := range r.values {
			if value.backingFile == "" {
				continue
			}
			if hash, ok := value.value.(*Hash); ok {
				for _, key := range hash.Keys() {
					stateValue := hash.Get(key)
					if stateValue.SaveNeeded() {
						json := stateValue.JSON(true)
						save(path.Join(r.basePath, "config", value.backingFile, key+".json"), json)
						stateValue.SetSaveNeeded(false)
					}
				}
			} else {
				if value.value.SaveNeeded() {
					json := value.value.JSON(true)
					save(path.Join(r.basePath, "config", value.backingFile+".json"), json)
					value.value.SetSaveNeeded(false)
				}
			}
		}
	}
	r.SetIsReady(true)
	for {
		saveAllNeeded()
		time.Sleep(1 * time.Second)
	}
}
