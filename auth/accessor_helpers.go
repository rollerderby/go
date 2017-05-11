package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"hash"
	"net/http"
)

func (h *RootUsers) AddUser(username, password string, isSuper bool, groups []string, personID string) (*User, error) {
	user, err := h.New(username)
	if err != nil {
		return nil, err
	}

	if err = user.SetUsername(username); err != nil {
		return user, err
	}
	if err = user.SetIsSuper(isSuper); err != nil {
		return user, err
	}
	if err = user.SetPersonID(personID); err != nil {
		return user, err
	}
	for _, group := range groups {
		if err := user.Groups().Add(group); err != nil {
			return user, err
		}
	}
	user.SetPassword(password)

	return user, nil
}

func (h *User) HasGroup(groups ...string) bool {
	if h.IsSuper() {
		return true
	}
	for _, g1 := range h.Groups().Values() {
		for _, g2 := range groups {
			if g1 == g2 {
				return true
			}
		}
	}
	return false
}

func (u *User) setCookie(w http.ResponseWriter) error {
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

func (h *User) CheckPassword(password string) bool {
	var hash hash.Hash

	switch h.PasswordHashType() {
	case "md5":
		hash = md5.New()
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	}

	if hash == nil {
		return false
	}

	hash.Write([]byte(password))
	passwordHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return passwordHash == h.PasswordHash()
}

func (h *User) SetPassword(password string) error {
	hash := sha512.New()
	hash.Write([]byte(password))
	passwordHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if err := h.SetPasswordHash(passwordHash); err != nil {
		return err
	}
	if err := h.SetPasswordHashType("sha512"); err != nil {
		return err
	}

	return nil
}
