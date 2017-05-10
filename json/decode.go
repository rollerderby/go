package json

import (
	"errors"
	"fmt"
)

func decodeArray(t *tokens) (Array, error) {
	var a Array
	for {
		tok := t.Next()
		if tok == nil {
			return nil, errors.New("JSON finished before done with array")
		}

		if tok.t == tokRightBracket {
			return a, nil
		}

		if len(a) == 0 {
			// should be a value
			val, err := decodeValue(t, tok)
			if err != nil {
				return nil, err
			}
			a = append(a, val)
		} else {
			// should be a comma followed by a value
			if tok.t != tokComma {
				return nil, fmt.Errorf("Expected Comma, got %v", tok)
			}
			val, err := decodeValue(t, nil)
			if err != nil {
				return nil, err
			}
			a = append(a, val)
		}
	}

}

func decodeObject(t *tokens) (Object, error) {
	o := make(Object)
	for {
		tok := t.Next()
		if tok == nil {
			return nil, errors.New("JSON finished before done with object")
		}

		switch tok.t {
		case tokRightBrace:
			return o, nil
		case tokString:
			key := tok.val

			tok = t.Next()
			if tok == nil || tok.t != tokColon {
				return nil, errors.New("Expected Colon")
			}
			val, err := decodeValue(t, nil)
			if err != nil {
				return nil, err
			}
			o[key] = val
		case tokComma:
			// Do Nothing
			break
		default:
			return nil, fmt.Errorf("Unexpected token: %v", tok)
		}
	}
}

func decodeValue(t *tokens, tok *token) (Value, error) {
	if tok == nil {
		tok = t.Next()
		if tok == nil {
			return nil, errors.New("JSON ended unexpected.  Looking for a value")
		}
	}

	switch tok.t {
	case tokTrue:
		return True, nil
	case tokFalse:
		return False, nil
	case tokNull:
		return Null, nil
	case tokString:
		return &String{val: tok.val}, nil
	case tokNumber:
		return &Number{val: tok.val}, nil
	case tokLeftBrace:
		return decodeObject(t)
	case tokLeftBracket:
		return decodeArray(t)
	}

	return nil, fmt.Errorf("Unexpected token: %v", tok)
}

func Decode(data []byte) (Value, error) {
	return decodeValue(newTokens(data), nil)
}
