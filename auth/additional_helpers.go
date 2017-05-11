package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
)

func (h *UsersHelper) AddUser(username, password string, isSuper bool, groups []string, personID string) (*UserHelper, error) {
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

func (h *UserHelper) HasGroup(groups ...string) bool {
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

func (h *UserHelper) CheckPassword(password string) bool {
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

func (h *UserHelper) SetPassword(password string) error {
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
