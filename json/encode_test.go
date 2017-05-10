package json

import "testing"

var encValue string

func TestEncoder(t *testing.T) {
	val, err := Decode([]byte(rawJSONShort))
	if err != nil {
		t.Fatal(err)
		return
	}

	encValue = val.JSON(false)
	if encValue != rawJSONShortClean {
		t.Fatalf("Unexected result.  `%v` != `%v`", encValue, rawJSONShortClean)
	}

	t.Log(encValue)
}

func BenchmarkEncoder(b *testing.B) {
	val, err := Decode(rawJSON)
	if err != nil {
		b.Fatal(err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encValue = val.JSON(false)
	}
}

/*
func BenchmarkEncoderParallel(b *testing.B) {
	val, err := Decode(rawJSON)
	if err != nil {
		b.Fatal(err)
		return
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			encValue = val.JSON(false)
		}
	})
}
*/
