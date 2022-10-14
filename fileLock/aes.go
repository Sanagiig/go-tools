package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func AesEncrypt(key, src []byte) (data []byte, err error) {

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	} else if len(src) == 0 {
		return nil, errors.New("src is empty")
	}

	plaintext, err := pkcs7Pad(src, block.BlockSize())

	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	bm := cipher.NewCBCEncrypter(block, iv)
	bm.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func AesDecrypt(key, src []byte) (data []byte, err error) {

	if len(src) < aes.BlockSize {
		return nil, errors.New("data length error")
	}

	iv := src[:aes.BlockSize]
	ciphertext := src[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bm := cipher.NewCBCDecrypter(block, iv)
	bm.CryptBlocks(ciphertext, ciphertext)
	ciphertext, err = pkcs7Unpad(ciphertext, aes.BlockSize)

	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func pkcs7Pad(src []byte, blockSize int) (dest []byte, err error) {
	if blockSize <= 0 {
		return nil, errors.New("block size is 0")
	} else if src == nil || len(src) == 0 {
		return nil, errors.New("src is nil")
	}
	n := blockSize - (len(src) % blockSize)
	pb := make([]byte, len(src)+n)
	copy(pb, src)
	copy(pb[len(src):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func pkcs7Unpad(src []byte, blockSize int) (dest []byte, err error) {

	if blockSize <= 0 {
		return nil, errors.New("block size is 0")
	} else if len(src)%blockSize != 0 {
		return nil, errors.New("src length error")
	} else if src == nil || len(src) == 0 {
		return nil, errors.New("src is nil")
	}

	c := src[len(src)-1]

	padLength := int(c)

	if padLength == 0 || padLength > len(src) {
		return nil, errors.New("pad length error")
	}

	for i := 0; i < padLength; i++ {
		if src[len(src)-padLength+i] != c {
			return nil, errors.New("pad content error")
		}
	}

	return src[:len(src)-padLength], nil

}
