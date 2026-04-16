package service

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"net/url"
	"sort"
	"strings"

	"github.com/dop251/goja"
)

// RegisterCryptoModule 向 goja VM 注册完整的 crypto 模块
//
// 注册内容分为两部分：
// 1. crypto 对象 — Node.js 风格链式调用 + 快捷函数 + AES/RSA
// 2. 顶层工具函数 — 签名场景常用快捷方式（无需 crypto. 前缀）
//
// ═══════════════════════════════════════════════════════════
// crypto 对象 API:
// ═══════════════════════════════════════════════════════════
//
// 【链式调用 — 兼容 Node.js crypto 习惯】
//   crypto.createHmac(algorithm, key).update(data).digest(encoding)
//   crypto.createHash(algorithm).update(data).digest(encoding)
//   algorithm: 'md5' | 'sha1' | 'sha256' | 'sha512'
//   encoding:  'hex' | 'base64'
//
// 【快捷哈希 — 返回 hex 字符串】
//   crypto.md5(data)
//   crypto.sha1(data)
//   crypto.sha256(data)
//   crypto.sha512(data)
//
// 【快捷 HMAC — 返回 hex 字符串】
//   crypto.hmac(algorithm, key, data)
//   crypto.hmacSha256(key, data) / crypto.hmacSha1 / crypto.hmacSha512 / crypto.hmacMd5
//
// 【编码工具】
//   crypto.base64Encode(str) / crypto.base64Decode(str)
//   crypto.hexEncode(str) / crypto.hexDecode(str)
//   crypto.urlEncode(str) / crypto.urlDecode(str)
//
// 【AES 对称加密/解密】
//   crypto.aesEncrypt(data, key, iv?)    — AES-CBC-PKCS7, 返回 base64
//   crypto.aesDecrypt(base64Str, key)    — AES-CBC-PKCS7 解密
//
// 【RSA 非对称签名/验签】
//   crypto.rsaSign(data, privateKeyPem, algorithm?)    — RSA 签名, 返回 base64
//   crypto.rsaVerify(data, signature, publicKeyPem, algorithm?) — 验签, 返回 bool
//   algorithm: 'sha1' | 'sha256' | 'sha512'
//
// 【签名辅助】
//   crypto.sortByKeys(obj)                     — 按 key 字典序排序对象
//   crypto.sortAndJoin(obj, separator?, kvSep?) — 排序后拼接字符串
//
// ═══════════════════════════════════════════════════════════
// 顶层快捷函数 (全局可用，无需 crypto. 前缀):
// ═══════════════════════════════════════════════════════════
//
//   hmacSha256(key, data) / hmacSha1 / hmacSha512 / hmacMd5
//   sha256(data) / sha1(data) / sha512(data) / md5(data)
//   base64Encode(str) / base64Decode(str)
//   hexEncode(str) / hexDecode(str)
//   urlEncode(str) / urlDecode(str)
//   sortByKeys(obj)
//   sortAndJoin(obj, '&', '=')
//
func RegisterCryptoModule(vm *goja.Runtime) {
	cryptoObj := vm.NewObject()

	registerChainedAPI(vm, cryptoObj)
	registerHashShortcuts(vm, cryptoObj)
	registerHmacShortcuts(vm, cryptoObj)
	registerEncodingTools(vm, cryptoObj)
	registerAES(vm, cryptoObj)
	registerRSA(vm, cryptoObj)
	registerSignatureHelpers(vm, cryptoObj)

	vm.Set("crypto", cryptoObj)

	registerTopLevelShortcuts(vm)
}

// ═══════════════════════════════════════════
// 链式调用 API: createHmac / createHash
// ═══════════════════════════════════════════

func registerChainedAPI(vm *goja.Runtime, cryptoObj *goja.Object) {
	cryptoObj.Set("createHmac", func(call goja.FunctionCall) goja.Value {
		algorithm := call.Argument(0).String()
		key := call.Argument(1).String()

		h, err := newHmac(algorithm, key)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}

		hmacObj := vm.NewObject()
		hmacObj.Set("update", func(call goja.FunctionCall) goja.Value {
			h.Write([]byte(call.Argument(0).String()))
			return hmacObj
		})
		hmacObj.Set("digest", func(call goja.FunctionCall) goja.Value {
			encoding := "hex"
			if len(call.Arguments) > 0 {
				encoding = call.Argument(0).String()
			}
			return vm.ToValue(encodeBytes(h.Sum(nil), encoding))
		})
		return hmacObj
	})

	cryptoObj.Set("createHash", func(call goja.FunctionCall) goja.Value {
		algorithm := call.Argument(0).String()

		h, err := newHash(algorithm)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}

		hashObj := vm.NewObject()
		hashObj.Set("update", func(call goja.FunctionCall) goja.Value {
			h.Write([]byte(call.Argument(0).String()))
			return hashObj
		})
		hashObj.Set("digest", func(call goja.FunctionCall) goja.Value {
			encoding := "hex"
			if len(call.Arguments) > 0 {
				encoding = call.Argument(0).String()
			}
			return vm.ToValue(encodeBytes(h.Sum(nil), encoding))
		})
		return hashObj
	})
}

// ═══════════════════════════════════════════
// 快捷哈希函数
// ═══════════════════════════════════════════

func registerHashShortcuts(vm *goja.Runtime, cryptoObj *goja.Object) {
	cryptoObj.Set("md5", func(call goja.FunctionCall) goja.Value {
		sum := md5.Sum([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	cryptoObj.Set("sha1", func(call goja.FunctionCall) goja.Value {
		sum := sha1.Sum([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	cryptoObj.Set("sha256", func(call goja.FunctionCall) goja.Value {
		sum := sha256.Sum256([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	cryptoObj.Set("sha512", func(call goja.FunctionCall) goja.Value {
		sum := sha512.Sum512([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
}

// ═══════════════════════════════════════════
// 快捷 HMAC 函数
// ═══════════════════════════════════════════

func registerHmacShortcuts(vm *goja.Runtime, cryptoObj *goja.Object) {
	// 通用: crypto.hmac(algorithm, key, data)
	cryptoObj.Set("hmac", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(vm.NewTypeError("crypto.hmac(algorithm, key, data) 需要3个参数"))
		}
		result, err := computeHmac(call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		return vm.ToValue(result)
	})

	cryptoObj.Set("hmacSha256", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha256", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	cryptoObj.Set("hmacSha1", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha1", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	cryptoObj.Set("hmacSha512", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha512", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	cryptoObj.Set("hmacMd5", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("md5", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
}

// ═══════════════════════════════════════════
// 编码工具
// ═══════════════════════════════════════════

func registerEncodingTools(vm *goja.Runtime, cryptoObj *goja.Object) {
	cryptoObj.Set("base64Encode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})
	cryptoObj.Set("base64Decode", func(call goja.FunctionCall) goja.Value {
		decoded, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("base64 decode failed: " + err.Error()))
		}
		return vm.ToValue(string(decoded))
	})
	cryptoObj.Set("hexEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(hex.EncodeToString([]byte(call.Argument(0).String())))
	})
	cryptoObj.Set("hexDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := hex.DecodeString(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("hex decode failed: " + err.Error()))
		}
		return vm.ToValue(string(decoded))
	})
	cryptoObj.Set("urlEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(url.QueryEscape(call.Argument(0).String()))
	})
	cryptoObj.Set("urlDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := url.QueryUnescape(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("url decode failed: " + err.Error()))
		}
		return vm.ToValue(decoded)
	})
}

// ═══════════════════════════════════════════
// AES 对称加密/解密
// ═══════════════════════════════════════════

func registerAES(vm *goja.Runtime, cryptoObj *goja.Object) {
	// crypto.aesEncrypt(data, key, iv?) — AES-CBC-PKCS7, 返回 base64
	cryptoObj.Set("aesEncrypt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("crypto.aesEncrypt(data, key, iv?) 需要2-3个参数"))
		}
		data := []byte(call.Argument(0).String())
		key := []byte(call.Argument(1).String())

		var iv []byte
		if len(call.Arguments) >= 3 && call.Argument(2).String() != "" {
			iv = []byte(call.Argument(2).String())
		} else {
			iv = make([]byte, aes.BlockSize)
			if _, err := rand.Read(iv); err != nil {
				panic(vm.NewTypeError("generate IV failed: " + err.Error()))
			}
		}

		if len(key) != 16 && len(key) != 24 && len(key) != 32 {
			panic(vm.NewTypeError(fmt.Sprintf("AES key 长度必须为 16/24/32 字节，当前: %d", len(key))))
		}
		if len(iv) != aes.BlockSize {
			panic(vm.NewTypeError(fmt.Sprintf("IV 长度必须为 %d 字节，当前: %d", aes.BlockSize, len(iv))))
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			panic(vm.NewTypeError("create AES cipher failed: " + err.Error()))
		}

		padded := pkcs7Pad(data, aes.BlockSize)
		encrypted := make([]byte, len(padded))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(encrypted, padded)

		result := append(iv, encrypted...)
		return vm.ToValue(base64.StdEncoding.EncodeToString(result))
	})

	// crypto.aesDecrypt(base64Str, key) — AES-CBC-PKCS7 解密
	cryptoObj.Set("aesDecrypt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("crypto.aesDecrypt(base64Str, key) 需要2个参数"))
		}

		encoded := call.Argument(0).String()
		key := []byte(call.Argument(1).String())

		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			panic(vm.NewTypeError("base64 decode failed: " + err.Error()))
		}

		if len(raw) < aes.BlockSize {
			panic(vm.NewTypeError("密文太短"))
		}

		iv := raw[:aes.BlockSize]
		ciphertext := raw[aes.BlockSize:]

		if len(key) != 16 && len(key) != 24 && len(key) != 32 {
			panic(vm.NewTypeError(fmt.Sprintf("AES key 长度必须为 16/24/32 字节，当前: %d", len(key))))
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			panic(vm.NewTypeError("create AES cipher failed: " + err.Error()))
		}

		if len(ciphertext)%aes.BlockSize != 0 {
			panic(vm.NewTypeError("密文长度不是块大小的整数倍"))
		}

		decrypted := make([]byte, len(ciphertext))
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(decrypted, ciphertext)

		plaintext, err := pkcs7Unpad(decrypted)
		if err != nil {
			panic(vm.NewTypeError("unpad failed: " + err.Error()))
		}

		return vm.ToValue(string(plaintext))
	})
}

// ═══════════════════════════════════════════
// RSA 签名/验签
// ═══════════════════════════════════════════

func registerRSA(vm *goja.Runtime, cryptoObj *goja.Object) {
	// crypto.rsaSign(data, privateKeyPem, algorithm?) — 返回 base64 签名
	cryptoObj.Set("rsaSign", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("crypto.rsaSign(data, privateKeyPem, algorithm?) 需要2-3个参数"))
		}

		data := []byte(call.Argument(0).String())
		privateKeyPem := call.Argument(1).String()
		algorithm := "sha256"
		if len(call.Arguments) >= 3 {
			algorithm = call.Argument(2).String()
		}

		key, err := parseRSAPrivateKey(privateKeyPem)
		if err != nil {
			panic(vm.NewTypeError("parse private key failed: " + err.Error()))
		}

		hasher, hashFunc, err := newHashForRSA(algorithm)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		hasher.Write(data)

		signature, err := rsa.SignPKCS1v15(rand.Reader, key, hashFunc, hasher.Sum(nil))
		if err != nil {
			panic(vm.NewTypeError("RSA sign failed: " + err.Error()))
		}

		return vm.ToValue(base64.StdEncoding.EncodeToString(signature))
	})

	// crypto.rsaVerify(data, signature, publicKeyPem, algorithm?) — 返回 bool
	cryptoObj.Set("rsaVerify", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(vm.NewTypeError("crypto.rsaVerify(data, signature, publicKeyPem, algorithm?) 需要3-4个参数"))
		}

		data := []byte(call.Argument(0).String())
		signatureBase64 := call.Argument(1).String()
		publicKeyPem := call.Argument(2).String()
		algorithm := "sha256"
		if len(call.Arguments) >= 4 {
			algorithm = call.Argument(3).String()
		}

		key, err := parseRSAPublicKey(publicKeyPem)
		if err != nil {
			panic(vm.NewTypeError("parse public key failed: " + err.Error()))
		}

		signature, err := base64.StdEncoding.DecodeString(signatureBase64)
		if err != nil {
			panic(vm.NewTypeError("signature base64 decode failed: " + err.Error()))
		}

		hasher, hashFunc, err := newHashForRSA(algorithm)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		hasher.Write(data)

		err = rsa.VerifyPKCS1v15(key, hashFunc, hasher.Sum(nil), signature)
		return vm.ToValue(err == nil)
	})
}

// ═══════════════════════════════════════════
// 签名辅助函数
// ═══════════════════════════════════════════

func registerSignatureHelpers(vm *goja.Runtime, cryptoObj *goja.Object) {
	// crypto.sortByKeys(obj) — 按 key 字典序排序，返回新对象
	cryptoObj.Set("sortByKeys", func(call goja.FunctionCall) goja.Value {
		obj := call.Argument(0).Export()
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			panic(vm.NewTypeError("sortByKeys 参数必须是对象"))
		}

		keys := make([]string, 0, len(objMap))
		for k := range objMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sorted := vm.NewObject()
		for _, k := range keys {
			sorted.Set(k, objMap[k])
		}
		return sorted
	})

	// crypto.sortAndJoin(obj, separator?, kvSeparator?)
	cryptoObj.Set("sortAndJoin", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("crypto.sortAndJoin(obj, separator?, kvSeparator?) 至少需要1个参数"))
		}

		obj := call.Argument(0).Export()
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			panic(vm.NewTypeError("第一个参数必须是对象"))
		}

		separator := "&"
		kvSeparator := "="
		if len(call.Arguments) >= 2 {
			separator = call.Argument(1).String()
		}
		if len(call.Arguments) >= 3 {
			kvSeparator = call.Argument(2).String()
		}

		keys := make([]string, 0, len(objMap))
		for k := range objMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+kvSeparator+fmt.Sprintf("%v", objMap[k]))
		}

		return vm.ToValue(strings.Join(parts, separator))
	})
}

// ═══════════════════════════════════════════
// 顶层快捷函数
// ═══════════════════════════════════════════

func registerTopLevelShortcuts(vm *goja.Runtime) {
	// HMAC 快捷
	vm.Set("hmacSha256", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha256", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	vm.Set("hmacSha1", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha1", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	vm.Set("hmacSha512", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("sha512", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})
	vm.Set("hmacMd5", func(call goja.FunctionCall) goja.Value {
		result, _ := computeHmac("md5", call.Argument(0).String(), call.Argument(1).String())
		return vm.ToValue(result)
	})

	// 哈希快捷
	vm.Set("sha256", func(call goja.FunctionCall) goja.Value {
		sum := sha256.Sum256([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	vm.Set("sha1", func(call goja.FunctionCall) goja.Value {
		sum := sha1.Sum([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	vm.Set("sha512", func(call goja.FunctionCall) goja.Value {
		sum := sha512.Sum512([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})
	vm.Set("md5", func(call goja.FunctionCall) goja.Value {
		sum := md5.Sum([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})

	// 编码快捷
	vm.Set("base64Encode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})
	vm.Set("base64Decode", func(call goja.FunctionCall) goja.Value {
		decoded, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("base64 decode failed: " + err.Error()))
		}
		return vm.ToValue(string(decoded))
	})
	vm.Set("hexEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(hex.EncodeToString([]byte(call.Argument(0).String())))
	})
	vm.Set("hexDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := hex.DecodeString(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("hex decode failed: " + err.Error()))
		}
		return vm.ToValue(string(decoded))
	})
	vm.Set("urlEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(url.QueryEscape(call.Argument(0).String()))
	})
	vm.Set("urlDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := url.QueryUnescape(call.Argument(0).String())
		if err != nil {
			panic(vm.NewTypeError("url decode failed: " + err.Error()))
		}
		return vm.ToValue(decoded)
	})

	// 签名辅助快捷
	vm.Set("sortByKeys", func(call goja.FunctionCall) goja.Value {
		obj := call.Argument(0).Export()
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			panic(vm.NewTypeError("sortByKeys 参数必须是对象"))
		}
		keys := make([]string, 0, len(objMap))
		for k := range objMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := vm.NewObject()
		for _, k := range keys {
			sorted.Set(k, objMap[k])
		}
		return sorted
	})

	vm.Set("sortAndJoin", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("sortAndJoin(obj, separator?, kvSeparator?) 至少需要1个参数"))
		}
		obj := call.Argument(0).Export()
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			panic(vm.NewTypeError("第一个参数必须是对象"))
		}

		separator := "&"
		kvSeparator := "="
		if len(call.Arguments) >= 2 {
			separator = call.Argument(1).String()
		}
		if len(call.Arguments) >= 3 {
			kvSeparator = call.Argument(2).String()
		}

		keys := make([]string, 0, len(objMap))
		for k := range objMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+kvSeparator+fmt.Sprintf("%v", objMap[k]))
		}
		return vm.ToValue(strings.Join(parts, separator))
	})
}

// ═══════════════════════════════════════════
// 内部工具函数
// ═══════════════════════════════════════════

func newHmac(algorithm, key string) (hash.Hash, error) {
	switch strings.ToLower(algorithm) {
	case "md5":
		return hmac.New(md5.New, []byte(key)), nil
	case "sha1":
		return hmac.New(sha1.New, []byte(key)), nil
	case "sha256":
		return hmac.New(sha256.New, []byte(key)), nil
	case "sha512":
		return hmac.New(sha512.New, []byte(key)), nil
	default:
		return nil, fmt.Errorf("不支持的 HMAC 算法: %s (支持: md5, sha1, sha256, sha512)", algorithm)
	}
}

func newHash(algorithm string) (hash.Hash, error) {
	switch strings.ToLower(algorithm) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %s (支持: md5, sha1, sha256, sha512)", algorithm)
	}
}

func computeHmac(algorithm, key, data string) (string, error) {
	h, err := newHmac(algorithm, key)
	if err != nil {
		return "", err
	}
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func encodeBytes(data []byte, encoding string) string {
	switch strings.ToLower(encoding) {
	case "base64":
		return base64.StdEncoding.EncodeToString(data)
	case "hex":
		return hex.EncodeToString(data)
	default:
		return hex.EncodeToString(data)
	}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > len(data) {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:len(data)-padding], nil
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM 块")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("不是 RSA 私钥")
		}
		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析 RSA 私钥失败 (PKCS1/PKCS8 均失败): %v", err)
	}
	return key, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM 块")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("不是 RSA 公钥")
		}
		return rsaKey, nil
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		rsaKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("证书中的公钥不是 RSA")
		}
		return rsaKey, nil
	}

	return nil, fmt.Errorf("解析 RSA 公钥失败")
}

func newHashForRSA(algorithm string) (hash.Hash, crypto.Hash, error) {
	switch strings.ToLower(algorithm) {
	case "sha1":
		return sha1.New(), crypto.SHA1, nil
	case "sha256":
		return sha256.New(), crypto.SHA256, nil
	case "sha512":
		return sha512.New(), crypto.SHA512, nil
	default:
		return nil, 0, fmt.Errorf("不支持的 RSA 签名算法: %s (支持: sha1, sha256, sha512)", algorithm)
	}
}
