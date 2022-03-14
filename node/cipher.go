package node

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// using algorithm:
//   alg( md5(hid+HexString(cipherKey)) + nanoTs , cipherKey)
func getAgentAuthKey(alg uint32, hid string, cipherKey []byte, nanoTs uint64) []byte {
	inner := fmt.Sprintf("%s%s", hid, hex.EncodeToString(cipherKey))
	digest := md5.Sum([]byte(inner))
	text := fmt.Sprintf("%s%d", hex.EncodeToString(digest[:]), nanoTs)

	if alg == AES { //CBC
		block, err := aes.NewCipher(cipherKey)
		if err != nil {
			log.Errorln("New cipher failed", err)
			return nil
		}

		blockSize := block.BlockSize()                                    // 获取秘钥块的长度
		padded := pkcs5Padding([]byte(text), blockSize)                   // 补全码
		blockMode := cipher.NewCBCEncrypter(block, cipherKey[:blockSize]) // 加密模式
		encrypted := make([]byte, len(padded))                            // 创建数组
		blockMode.CryptBlocks(encrypted, padded)                          // 加密
		return encrypted
	} else {
		log.Errorln("algorithm not supported yet")
		return nil
	}
}

// using algorithm:
//  alg(md5(id+hid+HexString(cipherKey)) + timestamp + expire)
func getServerAuthKey(alg uint32, id, hid string, cipherKey []byte, clientTs, expire uint64) []byte {
	inner := fmt.Sprintf("%s%s%s", id, hid, hex.EncodeToString(cipherKey))
	digest := md5.Sum([]byte(inner))
	text := fmt.Sprintf("%s%d%d", hex.EncodeToString(digest[:]), clientTs, expire)

	if alg == AES { //CBC
		block, err := aes.NewCipher(cipherKey)
		if err != nil {
			log.Errorln("New cipher failed", err)
			return nil
		}

		blockSize := block.BlockSize()                                    // 获取秘钥块的长度
		padded := pkcs5Padding([]byte(text), blockSize)                   // 补全码
		blockMode := cipher.NewCBCEncrypter(block, cipherKey[:blockSize]) // 加密模式
		encrypted := make([]byte, len(padded))                            // 创建数组
		blockMode.CryptBlocks(encrypted, padded)                          // 加密
		return encrypted
	} else {
		log.Errorln("algorithm not supported yet")
		return nil
	}
}

func getCipherKey(alg uint32) []byte {
	if alg == AES {
		str := fmt.Sprintf("%s++%s", "ooEyqiZ", "KdCIOlk")
		return []byte(str)
	} else if alg == DES {
		return []byte("E5QI2exh")
	} else {
		return []byte(fmt.Sprintf("%s%s%s%s%s%s%s%s",
			"2D10", "E80B", "ED23", "71EE", "0C4A", "0742", "B09D", "B296"))
	}
}
