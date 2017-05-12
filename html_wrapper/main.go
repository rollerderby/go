package html_wrapper

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/rollerderby/go/json"
	"github.com/rollerderby/go/logger"
)

var log = logger.New("html_wrapper")

type FileCache struct {
	sync.Mutex
	path  string
	cache map[string]*wrappedFile
}

type wrappedFile struct {
	path string
	name string
	fi   os.FileInfo
	buf  []byte
}

type File struct {
	r  *bytes.Reader
	fi os.FileInfo
}

func New(path string) *FileCache {
	return &FileCache{
		path:  path,
		cache: make(map[string]*wrappedFile),
	}
}

func (fc *FileCache) Open(name string) (http.File, error) {
	lcName := strings.ToLower(name)
	if strings.HasSuffix(lcName, ".wrapper") {
		return nil, os.ErrNotExist
	}
	if !strings.HasSuffix(lcName, ".html") {
		return os.Open(path.Join(fc.path, name))
	}

	f, err := fc.wrapFile(fc.path, name)
	return f, err
}

func (fc *FileCache) wrapFile(path, name string) (http.File, error) {
	fc.Lock()
	defer fc.Unlock()

	var err error

	f := fc.cache[name]
	if f == nil {
		f, err = newWrappedFile(path, name)
		if err != nil {
			return nil, err
		}
		fc.cache[name] = f
	} else {
		err = f.update()
		if err != nil {
			return nil, err
		}
	}

	return newFile(f.buf, f.fi), nil
}

func newWrappedFile(path, name string) (*wrappedFile, error) {
	wf := &wrappedFile{path: path, name: name}
	if err := wf.update(); err != nil {
		return nil, err
	}
	return wf, nil
}

func (wf *wrappedFile) update() error {
	fi, err := os.Stat(path.Join(wf.path, wf.name))
	if err != nil {
		return err
	}
	if wf.fi == nil || fi.Size() != wf.fi.Size() || fi.ModTime() != wf.fi.ModTime() {
		if err = wf.wrap(); err != nil {
			return err
		}

		wf.fi = fi
	}

	return nil
}

func (wf *wrappedFile) wrap() error {
	buf, err := ioutil.ReadFile(path.Join(wf.path, wf.name))
	if err != nil {
		return err
	}

	defer func() {
		wf.buf = buf
	}()

	contents := strings.TrimSpace(string(buf))
	if strings.HasPrefix(contents, "<!-- {") {
		endPos := strings.Index(contents, "} -->")
		if endPos != -1 {
			jsonStr := contents[5 : endPos+1]
			contentStr := strings.TrimSpace(contents[endPos+5:])
			if jValue, err := json.Decode([]byte(jsonStr)); err != nil {
				return fmt.Errorf("%q: Cannot decode wrapper config: %v", wf.name, err)
			} else {
				if wrappedContent, err := wrap(jValue, wf.path, contentStr); err != nil {
					return fmt.Errorf("%q: Cannot wrap file: %v", wf.name, err)
				} else {
					buf = []byte(wrappedContent)
					return nil
				}
			}
		}
	}
	return nil
}

func wrap(jConfig json.Value, rootPath, contents string) (string, error) {
	config, ok := jConfig.(json.Object)
	if !ok {
		return "", errors.New("JSON config is not an object")
	}

	template, title, javascript, css := "", "", []string{}, []string{}

	if val, ok := config["template"].(*json.String); ok {
		template = val.Get()
	}
	if val, ok := config["title"].(*json.String); ok {
		title = val.Get()
	}
	if val, ok := config["javascript"].(*json.String); ok {
		javascript = []string{fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`, val.Get())}
	} else if val, ok := config["javascript"].(json.Array); ok {
		for _, val := range val {
			if val, ok := val.(*json.String); ok {
				javascript = append(javascript, fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`, val.Get()))
			}
		}
	}
	if val, ok := config["css"].(*json.String); ok {
		css = []string{fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`, val.Get())}
	} else if val, ok := config["css"].(json.Array); ok {
		for _, val := range val {
			if val, ok := val.(*json.String); ok {
				css = append(css, fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`, val.Get()))
			}
		}
	}

	if template == "" {
		return "", errors.New("template is not specified")
	}

	wrapperBytes, err := ioutil.ReadFile(path.Join(rootPath, template+".wrapper"))
	if err != nil {
		return "", err
	}

	wrapper := string(wrapperBytes)

	wrapper = strings.Replace(wrapper, "%%title%%", title, -1)
	wrapper = strings.Replace(wrapper, "%%css%%", strings.Join(css, ""), -1)
	wrapper = strings.Replace(wrapper, "%%javascript%%", strings.Join(javascript, ""), -1)
	wrapper = strings.Replace(wrapper, "%%content%%", contents, -1)

	return wrapper, nil
}

func newFile(buf []byte, fi os.FileInfo) *File {
	return &File{r: bytes.NewReader(buf), fi: fi}
}

// http.File functions
func (f *File) Readdir(count int) ([]os.FileInfo, error) { return nil, os.ErrInvalid }
func (f *File) Stat() (os.FileInfo, error)               { return f, nil }

// io.Closer functions
func (f *File) Close() error { return nil }

// io.Reader functions
func (f *File) Read(b []byte) (n int, err error) { return f.r.Read(b) }

// io.Seeker functions
func (f *File) Seek(offset int64, whence int) (int64, error) { return f.r.Seek(offset, whence) }

// os.FileInfo functions
func (f *File) Name() string       { return f.fi.Name() }
func (f *File) Mode() os.FileMode  { return f.fi.Mode() }
func (f *File) ModTime() time.Time { return f.fi.ModTime() }
func (f *File) IsDir() bool        { return f.fi.IsDir() }
func (f *File) Size() int64        { return f.r.Size() }
func (f *File) Sys() interface{}   { return nil }
