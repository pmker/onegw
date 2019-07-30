# Openssl encryption

[![Build Status](https://www.travis-ci.org/thinkoner/openssl.svg?branch=master)](https://www.travis-ci.org/thinkoner/openssl)
[![Coverage Status](https://coveralls.io/repos/github/thinkoner/openssl/badge.svg?branch=master)](https://coveralls.io/github/thinkoner/openssl?branch=master)

A functions wrapping of OpenSSL library for symmetric and asymmetric encryption and decryption

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl/)

```
go get -u github.com/thinkoner/openssl
```

## Usage

### AES

The length of the key can be 16/24/32 characters (128/192/256 bits)

AES-ECB:

```go 
src := []byte("123456")
key := []byte("1234567890123456")
dst , _ := openssl.AesECBEncrypt(src, key, openssl.PKCS7_PADDING)
fmt.Printf(base64.StdEncoding.EncodeToString(dst))  // yXVUkR45PFz0UfpbDB8/ew==

dst , _ = openssl.AesECBDecrypt(dst, key, openssl.PKCS7_PADDING)
fmt.Println(string(dst)) // 123456
```

AES-CBC:

```go
src := []byte("123456")
key := []byte("1234567890123456")
iv := []byte("1234567890123456")
dst , _ := openssl.AesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
fmt.Println(base64.StdEncoding.EncodeToString(dst)) // 1jdzWuniG6UMtoa3T6uNLA==

dst , _ = openssl.AesCBCDecrypt(dst, key, iv, openssl.PKCS7_PADDING)
fmt.Println(string(dst)) // 123456
```

### DES

The length of the key must be 8 characters (64 bits).

DES-ECB:

```go
openssl.DesECBEncrypt(src, key, openssl.PKCS7_PADDING)
openssl.DesECBDecrypt(src, key, openssl.PKCS7_PADDING)
```

DES-CBC:

```go
openssl.DesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
openssl.DesCBCDecrypt(src, key, iv, openssl.PKCS7_PADDING)
```

### 3DES

The length of the key must be 24 characters (192 bits).

3DES-ECB:

```go
openssl.Des3ECBEncrypt(src, key, openssl.PKCS7_PADDING)
openssl.Des3ECBDecrypt(src, key, openssl.PKCS7_PADDING)
```

3DES-CBC:

```go
openssl.Des3CBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
openssl.Des3CBCDecrypt(src, key, iv, openssl.PKCS7_PADDING)
```

## License

This project is licensed under the [Apache 2.0 license](LICENSE).

## Contact

If you have any issues or feature requests, please contact us. PR is welcomed.
- https://github.com/thinkoner/openssl/issues
- duanpier@gmail.com

