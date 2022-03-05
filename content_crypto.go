package request

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/gildas/go-errors"
)

type CryptoAlgorithm uint
const (
	AESCTR CryptoAlgorithm = iota
)

func (algorithm CryptoAlgorithm) String() string {
	algorithms := [...]string{"AESCTR"}
	if int(algorithm) > len(algorithms) {
		return fmt.Sprintf("Unknown %d", algorithm)
	}
	return [...]string{"AESCTR"}[algorithm]
}

func (content Content) Decrypt(algorithm CryptoAlgorithm, key []byte) (*Content, error) {
	switch algorithm {
	case AESCTR:
		return content.DecryptWithAESCTR(key)
	}
	return nil, errors.InvalidType.With(algorithm.String())
}

func (content Content) DecryptWithAESCTR(key []byte) (*Content, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.ArgumentInvalid.With("key", key).(errors.Error).Wrap(err)
	}

	decrypted := Content{
		Type:    content.Type,
		Name:    content.Name,
		URL:     content.URL,
		Headers: content.Headers,
		Cookies: content.Cookies,
		Length:  content.Length,
		Data:    make([]byte, len(content.Data)),
	}

	stream := cipher.NewCTR(block, make([]byte, aes.BlockSize))
	stream.XORKeyStream(decrypted.Data, content.Data)
	return &decrypted, nil
}

func (content Content) Encrypt(algorithm CryptoAlgorithm, key []byte) (*Content, error) {
	switch algorithm {
	case AESCTR:
		return content.EncryptWithAESCTR(key)
	}
	return nil, errors.InvalidType.With(algorithm.String())
}

func (content Content) EncryptWithAESCTR(key []byte) (*Content, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.ArgumentInvalid.With("key", key).(errors.Error).Wrap(err)
	}

	encrypted := Content{
		Type:    content.Type,
		Name:    content.Name,
		URL:     content.URL,
		Headers: content.Headers,
		Cookies: content.Cookies,
		Length:  content.Length,
		Data:    make([]byte, len(content.Data)),
	}

	stream := cipher.NewCTR(block, make([]byte, aes.BlockSize))
	stream.XORKeyStream(encrypted.Data, content.Data)
	return &encrypted, nil
}
