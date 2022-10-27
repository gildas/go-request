package request

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"

	"github.com/gildas/go-errors"
)

type CryptoAlgorithm uint

const (
	NONE CryptoAlgorithm = iota
	AESCTR
)

func (algorithm CryptoAlgorithm) String() string {
	algorithms := [...]string{"NONE", "AESCTR"}
	if int(algorithm) > len(algorithms) {
		return fmt.Sprintf("Unknown %d", algorithm)
	}
	return algorithms[algorithm]
}

func CryptoAlgorithmFromString(algorithm string) (CryptoAlgorithm, error) {
	switch algorithm {
	case "NONE":
		return NONE, nil
	case "AESCTR":
		return AESCTR, nil
	}
	return NONE, errors.ArgumentInvalid.With("algorithm", algorithm)
}

func (algorithm CryptoAlgorithm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", algorithm.String())), nil
}

func (algorithm *CryptoAlgorithm) UnmarshalJSON(data []byte) (err error) {
	var value string
	if err = json.Unmarshal(data, &value); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*algorithm, err = CryptoAlgorithmFromString(value)
	return errors.JSONUnmarshalError.Wrap(err)
}

func (content Content) Decrypt(algorithm CryptoAlgorithm, key []byte) (*Content, error) {
	switch algorithm {
	case NONE:
		return &content, nil
	case AESCTR:
		return content.DecryptWithAESCTR(key)
	}
	return nil, errors.InvalidType.With(algorithm.String())
}

func (content Content) DecryptWithAESCTR(key []byte) (*Content, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WrapErrors(errors.ArgumentInvalid.With("key", key), err)
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
	case NONE:
		return &content, nil
	case AESCTR:
		return content.EncryptWithAESCTR(key)
	}
	return nil, errors.InvalidType.With(algorithm.String())
}

func (content Content) EncryptWithAESCTR(key []byte) (*Content, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WrapErrors(errors.ArgumentInvalid.With("key", key), err)
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
