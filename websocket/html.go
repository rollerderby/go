package websocket

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Time{}
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/websocket.js": {
		local: "../ws/html/websocket.js",
		size:  5267,
		compressed: `
H4sIAAAJbogA/+xYTW/jNhM+27+C0WEjvzGUvIcWhdygKII9pNjdAMmhh8UioGXaVqKQAiknDQr9987w
Wx9x4rYIUGAPm7XImYfzPDMaknqkkjwpck7+nE5E3ZSCqxx/T7ZCNZw+sJxUoqA4kbmhOUzXQjbRFD7i
8Iotd5ucrGmltBndNeKaFYJzVoB9I3fGu+SbS94w+UirnPxE/kd+PIM//z87O8Np6TyCzQ96poV/ShT3
DLD4rqrgsaBVtaTFPYTN2RP5VUr6nM5gwoPAWiGinayc63RyyUsAWu94gSRSmJsTEGGmBVgLSVKUBymT
koeZyZPKrFZfcfIbyAfP5vcCDNqptgE8mIG/C1hropink8LkLWowJwEqi0WZLUY8PKOO20CsmV7OOABE
ilgt0r31efCUDaFyTXABoyw5OtcCzZCpZM1OcgwGlVhSxW4Nq+RJqfz0NCEncSyuQmA0jZlBdZBfyHF+
3DXX4zk5PtZ0QwTnaHT7QO/ZjR5I/cLaHX4ETsblcvVHTrBCIq9uZg1TZIEepuAn/WJyAxru5ASj8AO4
2mQylNBjex0dPf0yuDn0VKJiWSU2adLll1nUZE4st4kpIhNPkAUr/He2tKoE28gqE1zUjIOxHjRPY1ZF
JRQLZvpxzO6BKUU3kaUdGLNlUsJb4y31o7ZrjXommkg81pOuW4LkwwfiBzOfCXJuV/AjXmRTr7p8BL+C
tWAFKydh0AGc3ekpuRK1OiLki2hIs2WEQkCPjBjIOTHqlA2hfEUgPdBFdowIbtxH9Y5fqnYxZugkf93S
Cfm6ZUjPa7Z69TSuLp8VPfVeabnQcfytvPy31O9qbE3fS+XPZrnv9f+W+tfLvVdmPuJi3/OyPy/4a2rt
7O7o9zurJs75Y4Dp9sMU7tmSBxuywYBN2OTGnuLicySywYPkYohsnP3hqX9QzDrbURvCvhhvvYfFrUHe
Gnj3ZDjkocH2EYkbeMTk80sN7jAuFuYANrUUBThd2JuAjWzAywLvY9ZtmhG3j+MN4jBmGuTfypIG28cl
bjOOyQ3jq4hF81yzOVnRhhqARj4bIDwm310t79wxeYKWcH1De/2MPrn2nLqXVvsogR3nt5urL5lqJFAq
188pIpnXe98BuS8YxgpimRgR2EC0lmin+2YKjFNv1BK4lhZb4vv4m9R7WT6/aJwRp2mv+oZFgsKI5d0i
FlhobbVONZX4NmU6Dd3opRwrpoJyDpuD9tMQOcGiktJVVnRra18q0rES7TEBUGEThxDwO8Nk9FRDdng1
cdNZIz6JJyYvqG/qaLKF/apiq1DhLi0NbptJLfgmcXkItvi9wCQg3iURj+I9/xPjm2Zrrov+O0BW6VGz
nfhLfAlWZwv47+fYFQZOTvwujXYOpgf6tfxmdy6M2Q0bPSD+xmP0greD3gFrI3WqusKym55T5MhCjL4Y
Fyb56xK2fxqiRZ54F3dZCGXrm5gu3Zeu/m/qYBohmfXqyl1O+53IfW1IhzX5D3aduHWGazG4LALPwz5z
mKVMeC9E1/mKhae/o17jDvH22rkv4Dd3vusIIJmHc2avEXXE9TmuzbeufbTPR2j7rNi2myYIlPiPLNds
U6qGycHWgY8GqfMG1ju1TTV+2De6fUHvIuhtIHF9u1o7/SsAAP//+uHjB5MUAAA=
`,
	},

	"/": {
		isDir: true,
		local: "../ws/html",
	},
}
