package entity

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"dragonfly-monitor/internal/common"
)

type DataSourceType int32

const (
	DataSourceTypeMongo DataSourceType = 1
	DataSourceTypeMysql DataSourceType = 2
)

// MonitorDatabase 被监控数据源
type MonitorDatabase struct {
	common.BaseEntity

	Database string         `json:"database" gorm:"column:database"`
	Name     string         `json:"name" gorm:"column:name"`
	Username string         `json:"username" gorm:"column:username"`
	Password string         `json:"password" gorm:"column:password"`
	Salt     string         `json:"salt" gorm:"column:salt"`
	Url      string         `json:"url" gorm:"column:url"`
	Type     DataSourceType `json:"type" gorm:"column:type"`
	Params   string         `json:"params" gorm:"column:params"`
}

func (m *MonitorDatabase) TableName() string {
	return "t_monitor_database"
}

// GetUrl 返回数据源连接地址
func (m *MonitorDatabase) GetUrl() string {
	return m.Url
}

// GetParams 返回数据源连接参数
func (m *MonitorDatabase) GetParams() string {
	return m.Params
}

// ResetEncodePasswordAndSalt 生成盐并加密密码
func (m *MonitorDatabase) ResetEncodePasswordAndSalt(plain string) error {
	salt, err := randomSalt(8)
	if err != nil {
		return err
	}
	m.Salt = salt
	enc, err := desEncrypt(plain, salt)
	if err != nil {
		return err
	}
	m.Password = enc
	return nil
}

// GetDecodePassword 解密密码
func (m *MonitorDatabase) GetDecodePassword() (string, error) {
	if m.Password == "" {
		return "", nil
	}
	if m.Salt == "" {
		return m.Password, nil
	}
	return desDecrypt(m.Password, m.Salt)
}

func randomSalt(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:n], nil
}

func desKeyFromSalt(salt string) []byte {
	key := []byte(salt)
	if len(key) < 8 {
		pad := make([]byte, 8)
		copy(pad, key)
		return pad
	}
	return key[:8]
}

func desEncrypt(plain, salt string) (string, error) {
	block, err := des.NewCipher(desKeyFromSalt(salt))
	if err != nil {
		return "", err
	}
	src := pkcs5Padding([]byte(plain), block.BlockSize())
	dst := make([]byte, len(src))
	// CBC with zero IV for simplicity and compatibility
	iv := make([]byte, block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(dst, src)
	return base64.StdEncoding.EncodeToString(dst), nil
}

func desDecrypt(cipherText, salt string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := des.NewCipher(desKeyFromSalt(salt))
	if err != nil {
		return "", err
	}
	if len(raw)%block.BlockSize() != 0 {
		return "", errors.New("cipher length invalid")
	}
	iv := make([]byte, block.BlockSize())
	mode := cipher.NewCBCDecrypter(block, iv)
	dst := make([]byte, len(raw))
	mode.CryptBlocks(dst, raw)
	out, err := pkcs5UnPadding(dst)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func pkcs5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(src, pad...)
}

func pkcs5UnPadding(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty")
	}
	padding := int(src[len(src)-1])
	if padding <= 0 || padding > len(src) {
		return nil, fmt.Errorf("bad padding")
	}
	return src[:len(src)-padding], nil
}
