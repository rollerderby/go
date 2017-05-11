package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
)

func Require(w *http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("auth")
	if err != nil {
		log.Error(err)
		return false
	}
	if c == nil {
		log.Notice("auth cookie not found.  redirecting to login page.")
		return false
	}
	return true
}

func verifySignature(data, signature []byte) error {
	hashed := sha256.Sum256(data)
	if err := rsa.VerifyPKCS1v15(&rsaKey.PublicKey, crypto.SHA256, hashed[:], signature); err != nil {
		return err
	}
	return nil
}

func sign(data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	if sig_bytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed[:]); err != nil {
		log.Errorf("Cannot generate auth signature.  %v", err)
		return nil, err
	} else {
		return sig_bytes, nil
	}
}

func CheckAuth(r *http.Request) *UserHelper {
	return loadAuthFromCookies(r, "auth")
}

func loadAuthFromCookies(r *http.Request, prefix string) *UserHelper {
	var auth, auth_sig *http.Cookie
	var sig_bytes []byte
	var err error

	authCookieName := prefix
	authCookieSigName := fmt.Sprintf("%v_sig", prefix)

	if auth, err = r.Cookie(authCookieName); err != nil {
		return nil
	}
	if auth_sig, err = r.Cookie(authCookieSigName); err != nil {
		return nil
	}

	if sig_bytes, err = base64.StdEncoding.DecodeString(auth_sig.Value); err != nil {
		log.Errorf("Cannot decode %v cookie.  %v", authCookieName, err)
		return nil
	}
	if err = verifySignature([]byte(auth.Value), sig_bytes); err != nil {
		log.Errorf("Cannot verify signature.  %v", err)
		return nil
	}

	users := Users.FindByUsername(auth.Value)
	if len(users) == 1 {
		return users[0]
	}
	return nil
}

func (u *UserHelper) setCookie(w http.ResponseWriter) error {
	setCookiePair := func(t, auth, sig string) {
		if auth != "" {
			http.SetCookie(w, &http.Cookie{Path: "/", Name: t, Value: auth, MaxAge: 0})
		} else {
			http.SetCookie(w, &http.Cookie{Path: "/", Name: t, MaxAge: -1})
		}
		if sig != "" {
			http.SetCookie(w, &http.Cookie{Path: "/", Name: t + "_sig", Value: sig, MaxAge: 0})
		} else {
			http.SetCookie(w, &http.Cookie{Path: "/", Name: t + "_sig", MaxAge: -1})
		}
	}

	var sig_bytes []byte
	var err error

	if sig_bytes, err = sign([]byte(u.Username())); err != nil {
		setCookiePair("auth", "", "")
		return errors.New("Cannot generate auth signature")
	}

	setCookiePair("auth", u.Username(), base64.StdEncoding.EncodeToString(sig_bytes))
	return nil
}

func Authenticate(w http.ResponseWriter, r *http.Request, username, password string) (*UserHelper, error) {
	user := CheckAuth(r)
	if user != nil {
		return user, nil
	}

	for _, user := range Users.FindByUsername(username) {
		if user.CheckPassword(password) {
			log.Infof("Login success: %q", username)
			if err := user.setCookie(w); err != nil {
				return nil, err
			}
			return user, nil
		} else {
			return nil, fmt.Errorf("Login failure, invalid password: %q", username)
		}
	}

	return nil, fmt.Errorf("Login failure, user not found: %q", username)
}
