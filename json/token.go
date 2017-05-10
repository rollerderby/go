package json

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type tokenType uint8

const (
	tokLeftBrace tokenType = iota
	tokRightBrace
	tokColon
	tokComma
	tokLeftBracket
	tokRightBracket
	tokString
	tokNumber
	tokTrue
	tokFalse
	tokNull
	tokError
)

type token struct {
	t   tokenType
	val string
	err error
}

type tokens struct {
	in io.RuneScanner
}

var tokenLeftBrace *token = &token{t: tokLeftBrace, val: "{"}
var tokenRightBrace *token = &token{t: tokRightBrace, val: "}"}
var tokenColon *token = &token{t: tokColon, val: ":"}
var tokenComma *token = &token{t: tokComma, val: ","}
var tokenLeftBracket *token = &token{t: tokLeftBracket, val: "["}
var tokenRightBracket *token = &token{t: tokRightBracket, val: "]"}
var tokenTrue *token = &token{t: tokTrue, val: "true"}
var tokenFalse *token = &token{t: tokFalse, val: "false"}
var tokenNull *token = &token{t: tokNull, val: "null"}

func (t *token) String() string {
	if t.t == tokError {
		return fmt.Sprintf("Error: %v", t.err)
	}
	return fmt.Sprintf("Token %v: %q", t.t, t.val)
}

func (t tokenType) String() string {
	switch t {
	case tokLeftBrace:
		return "LeftBrace"
	case tokRightBrace:
		return "RightBrace"
	case tokColon:
		return "Colon"
	case tokComma:
		return "Comma"
	case tokLeftBracket:
		return "LeftBracket"
	case tokRightBracket:
		return "RightBracket"
	case tokString:
		return "String"
	case tokNumber:
		return "Number"
	case tokTrue:
		return "True"
	case tokFalse:
		return "False"
	case tokNull:
		return "Null"
	case tokError:
		return "Error"
	}
	return fmt.Sprintf("UNKNOWN(%v)", uint8(t))
}

type tokensList struct {
	tokens []*token
	pos    int
}

func (t *tokensList) Next() *token {
	if t.pos == len(t.tokens) {
		return nil
	}
	t.pos++
	return t.tokens[t.pos-1]
}

func (t *tokens) tokenizeString() *token {
	val := GetBuffer()

	for {
		r, _, err := t.in.ReadRune()
		if err != nil {
			val.Return()
			return &token{t: tokError, err: err}
		}

		if r == '"' {
			tok := &token{t: tokString, val: val.String()}
			val.Return()
			return tok
		}

		if r == '\\' {
			r, _, err := t.in.ReadRune()
			if err != nil {
				val.Return()
				return &token{t: tokError, err: err}
			}
			switch r {
			case '"':
				val.WriteRune('"')
			case '\\':
				val.WriteRune('\\')
			case '/':
				val.WriteRune('/')
			case 'b':
				val.WriteRune('\b')
			case 'f':
				val.WriteRune('\f')
			case 'n':
				val.WriteRune('\n')
			case 't':
				val.WriteRune('\t')
			case 'r':
				val.WriteRune('\r')
			}
		} else {
			val.WriteRune(r)
		}
	}
}

func (t *tokens) tokenizeNumber() (string, error) {
	r, _, err := t.in.ReadRune()
	if err != nil {
		return "", err
	}

	val := GetBuffer()

	val.WriteRune(r)

	for {
		r, _, err := t.in.ReadRune()
		if err != nil {
			ret := val.String()
			val.Return()
			return ret, err
		}

		if r == '-' || r == 'e' || r == 'E' || r == '.' || r == '+' || unicode.IsDigit(r) {
			val.WriteRune(r)
		} else {
			break
		}
	}

	t.in.UnreadRune()

	ret := val.String()
	val.Return()
	return ret, nil
}

func (t *tokens) tokenizeExact(lookFor string, ignoreCase bool) error {
	for _, f := range lookFor {
		r, _, err := t.in.ReadRune()
		if err != nil {
			return err
		}

		t := r
		if ignoreCase {
			f = unicode.ToLower(f)
			t = unicode.ToLower(t)
		}

		if t != f {
			return fmt.Errorf("Cannot fully parse %q", lookFor)
		}
	}

	return nil
}

func newTokens(data []byte) *tokens {
	return &tokens{
		in: bytes.NewBuffer(data),
	}
}

func (t *tokens) Next() *token {
	for {
		r, _, err := t.in.ReadRune()
		if err != nil {
			if err != io.EOF {
				return &token{t: tokError, err: err}
			}
			return nil
		}

		switch r {
		case '{':
			return tokenLeftBrace
		case '}':
			return tokenRightBrace
		case ':':
			return tokenColon
		case ',':
			return tokenComma
		case '[':
			return tokenLeftBracket
		case ']':
			return tokenRightBracket
		case '"':
			return t.tokenizeString()
		default:
			if unicode.IsDigit(r) || r == '-' {
				// Number!
				t.in.UnreadRune()
				val, err := t.tokenizeNumber()
				if err != nil {
					if err != io.EOF {
						return &token{t: tokError, err: err}
					}
					return nil
				}
				return &token{t: tokNumber, val: val}
			} else if unicode.ToLower(r) == 't' {
				t.in.UnreadRune()
				err := t.tokenizeExact("true", true)
				if err != nil {
					if err != io.EOF {
						return &token{t: tokError, err: err}
					}
					return nil
				}
				return tokenTrue
			} else if unicode.ToLower(r) == 'f' {
				t.in.UnreadRune()
				err := t.tokenizeExact("false", true)
				if err != nil {
					if err != io.EOF {
						return &token{t: tokError, err: err}
					}
					return nil
				}
				return tokenFalse
			} else if unicode.ToLower(r) == 'n' {
				t.in.UnreadRune()
				err := t.tokenizeExact("null", true)
				if err != nil {
					if err != io.EOF {
						return &token{t: tokError, err: err}
					}
					return nil
				}
				return tokenNull
			}
		}
	}
}
