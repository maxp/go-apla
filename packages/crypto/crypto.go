package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	log "github.com/sirupsen/logrus"
)

// TODO In order to add new crypto provider with another key length it will be neccecary to fix constant blocksizes like
// crypto func getSharedKey() pub.X = new(big.Int).SetBytes(public[0:32])
// egcons func checkKey() gSettings.Key = hex.EncodeToString(privKey[aes.BlockSize:])

type cryptoProvider int
type ellipticSizeProvider int

const (
	_AESCBC cryptoProvider = iota
)

const (
	elliptic256 ellipticSizeProvider = iota
)

var (
	HashingError           = errors.New("Hashing error")
	EncryptingError        = errors.New("Encoding error")
	DecryptingError        = errors.New("Decrypting error")
	UnknownProviderError   = errors.New("Unknown provider")
	HashingEmpty           = errors.New("Hashing empty value")
	EncryptingEmpty        = errors.New("Encrypting empty value")
	DecryptingEmpty        = errors.New("Decrypting empty value")
	SigningEmpty           = errors.New("Signing empty value")
	CheckingSignEmpty      = errors.New("Cheking sign of empty")
	IncorrectSign          = errors.New("Incorrect sign")
	UnsupportedCurveSize   = errors.New("Unsupported curve size")
	IncorrectPrivKeyLength = errors.New("Incorrect private key length")
	IncorrectPubKeyLength  = errors.New("Incorrect public key length")
)

var (
	cryptoProv   = _AESCBC
	hashProv     = _SHA256
	ellipticSize = elliptic256
	signProv     = _ECDSA
	checksumProv = _CRC64
)

func Encrypt(msg []byte, key []byte, iv []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Debug(EncryptingEmpty.Error())
	}
	switch cryptoProv {
	case _AESCBC:
		return encryptCBC(msg, key, iv)
	default:
		return nil, UnknownProviderError
	}
}

func Decrypt(msg []byte, key []byte, iv []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Debug(DecryptingEmpty.Error())
	}
	switch cryptoProv {
	case _AESCBC:
		return decryptCBC(msg, key, iv)
	default:
		return nil, UnknownProviderError
	}
}

// SharedEncrypt creates a shared key and encrypts text. The first 32 characters are the created public key.
// The cipher text can be only decrypted with the original private key.
func SharedEncrypt(public, text []byte) ([]byte, error) {
	priv, pub, err := GenBytesKeys()
	if err != nil {
		return nil, err
	}
	shared, err := getSharedKey(priv, public)
	if err != nil {
		return nil, err
	}
	val, err := Encrypt(shared, text, pub)
	return val, err
}

// GenBytesKeys generates a random pair of ECDSA private and public binary keys.
func GenBytesKeys() ([]byte, []byte, error) {
	var curve elliptic.Curve
	switch ellipticSize {
	case elliptic256:
		curve = elliptic.P256()
	default:
		return nil, nil, UnsupportedCurveSize
	}
	private, err := ecdsa.GenerateKey(curve, crand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return private.D.Bytes(), append(converter.FillLeft(private.PublicKey.X.Bytes()), converter.FillLeft(private.PublicKey.Y.Bytes())...), nil
}

// GenHexKeys generates a random pair of ECDSA private and public hex keys.
func GenHexKeys() (string, string, error) {
	priv, pub, err := GenBytesKeys()
	if err != nil {
		return ``, ``, err
	}
	return hex.EncodeToString(priv), hex.EncodeToString(pub), nil
}

// CBCEncrypt encrypts the text by using the key parameter. It uses CBC mode of AES.
func encryptCBC(text, key, iv []byte) ([]byte, error) {
	if iv == nil {
		iv = make([]byte, consts.BlockSize)
		if _, err := crand.Read(iv); err != nil {
			return nil, err
		}
	} else if len(iv) < consts.BlockSize {
		return nil, fmt.Errorf(`wrong size of iv %d`, len(iv))
	} else if len(iv) > consts.BlockSize {
		iv = iv[:consts.BlockSize]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := _PKCS7Padding(text, consts.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plaintext))
	mode.CryptBlocks(encrypted, plaintext)
	return append(iv, encrypted...), nil

}

// CBCDecrypt decrypts the text by using key. It uses CBC mode of AES.
func decryptCBC(ciphertext, key, iv []byte) ([]byte, error) {
	if iv == nil {
		iv = ciphertext[:consts.BlockSize]
		ciphertext = ciphertext[consts.BlockSize:]
	}
	if len(ciphertext) < consts.BlockSize || len(ciphertext)%consts.BlockSize != 0 {
		return nil, fmt.Errorf(`Wrong size of cipher %d`, len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ret := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv[:consts.BlockSize]).CryptBlocks(ret, ciphertext)
	if ret, err = _PKCS7UnPadding(ret); err != nil {
		return nil, err
	}
	return ret, nil

}

// PKCS7Padding realizes PKCS#7 encoding which is described in RFC 5652.
func _PKCS7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	return append(src, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// PKCS7UnPadding realizes PKCS#7 decoding.
func _PKCS7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	padLength := int(src[length-1])
	for i := length - padLength; i < length; i++ {
		if int(src[i]) != padLength {
			return nil, fmt.Errorf(`incorrect input of PKCS7UnPadding`)
		}
	}
	return src[:length-int(src[length-1])], nil

}

// GetSharedKey creates and returns the shared key = private * public.
// public must be the public key from the different private key.
func getSharedKey(private, public []byte) (shared []byte, err error) {
	var pubkeyCurve elliptic.Curve
	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		return nil, UnknownProviderError
	}

	switch signProv {
	case _ECDSA:
		if len(private) != consts.PubkeySizeLength/2 {
			return nil, IncorrectPrivKeyLength
		}
		if len(public) != consts.PubkeySizeLength {
			return nil, IncorrectPubKeyLength
		}

		pub := new(ecdsa.PublicKey)
		pub.Curve = pubkeyCurve
		pub.X = new(big.Int).SetBytes(public[0 : consts.PubkeySizeLength/2])
		pub.Y = new(big.Int).SetBytes(public[consts.PubkeySizeLength/2:])

		bi := new(big.Int).SetBytes(private)
		priv := new(ecdsa.PrivateKey)
		priv.PublicKey.Curve = pubkeyCurve
		priv.D = bi
		priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())

		if priv.Curve.IsOnCurve(pub.X, pub.Y) {
			x, y := pub.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
			bytes := x.Bytes()
			bytes = append(bytes, y.Bytes()...)
			key, err := Hash(bytes)
			if err != nil {
				return nil, UnknownProviderError
			}
			shared = key
		} else {
			err = fmt.Errorf("Not IsOnCurve")
		}
	default:
		return nil, UnknownProviderError
	}

	return
}
