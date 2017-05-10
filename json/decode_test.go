package json

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestDecoder(t *testing.T) {
	obj1, err := Decode(rawJSON)
	if err != nil {
		t.Fatal(err)
		return
	}
	json1 := obj1.JSON(false)

	obj2, err := Decode([]byte(json1))
	if err != nil {
		t.Fatal(err)
		return
	}
	json2 := obj2.JSON(false)

	if json1 != json2 {
		t.Fatal("json1 != json2")
		return
	}
}

var benchValue Value

func BenchmarkDecoder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val, err := Decode(rawJSON)
		if err != nil {
			b.Fatal(err)
			return
		}
		benchValue = val
	}
}

func BenchmarkDecoderParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val, err := Decode(rawJSON)
			if err != nil {
				b.Fatal(err)
				return
			}
			benchValue = val
		}
	})
}

var rawJSON []byte
var rawJSONShort string = `{"test": 1, "boo":"hello"}`
var rawJSONShortClean string = `{"boo": "hello", "test": 1}`

func TestMain(m *testing.M) {
	url := "https://www.googleapis.com/discovery/v1/apis/admin/directory_v1/rest"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	rawJSON, err = ioutil.ReadAll(resp.Body)

	os.Exit(m.Run())
}
