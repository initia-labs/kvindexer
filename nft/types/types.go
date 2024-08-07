package types

import (
	"errors"
	"strconv"
)

func (m *TokenHandle) AdjustLength(delta int64) error {
	if m == nil {
		return errors.New("TokenHandle is nil")
	}

	i, err := strconv.ParseInt(m.Length, 10, 64)
	if err != nil && m.Length != "" {
		return err
	}
	i += delta
	if i < 0 {
		return errors.New("TokenHandle length cannot be negative")
	}

	m.Length = strconv.FormatInt(i, 10)
	return nil
}

func (m *Collection) AdjustLength(delta int64) error {
	if m == nil {
		return errors.New("Collection is nil")
	}
	if m.Nfts == nil {
		m.Nfts = &TokenHandle{}
	}
	return m.Nfts.AdjustLength(delta)
}

func (m *IndexedCollection) AdjustLength(delta int64) error {
	if m == nil {
		return errors.New("IndexedCollection is nil")
	}
	if m.Collection == nil {
		m.Collection = &Collection{}
	}
	return m.Collection.AdjustLength(delta)
}
