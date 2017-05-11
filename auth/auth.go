package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
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

func CheckAuth(r *http.Request) *User {
	return loadAuthFromCookies(r, "auth")
}

func loadAuthFromCookies(r *http.Request, prefix string) *User {
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

func Authenticate(w http.ResponseWriter, r *http.Request, username, password string) (*User, error) {
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
