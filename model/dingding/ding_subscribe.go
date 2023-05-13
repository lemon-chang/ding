package dingding

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"ding/model/common"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	r "math/rand"
	"sort"
	"strings"
	"time"
)

type DingSubscribe struct {
	EventJson map[string]interface{}
}

func NewDingSubscribe(EventJson map[string]interface{}) *DingSubscribe {
	return &DingSubscribe{EventJson: EventJson}
}

// CheckIn 用户签到事件
func (s *DingSubscribe) CheckIn(c *gin.Context) {
	var checkIn string = "用户签到事件触发" + fmt.Sprintf("%v;%v;%v;%v", s.EventJson["EventType"], s.EventJson["TimeStamp"], s.EventJson["CorpId"], s.EventJson["StaffId"])
	p := ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{
				IsAtAll: true,
			},
			Text: common.Text{
				Content: checkIn,
			},
			Msgtype: "text",
		},
		RepeatTime: "立即发送",
		TaskName:   "事件订阅",
	}

	(&DingRobot{RobotId: "2e36bf946609cd77206a01825273b2f3f33aed05eebe39c9cc9b6f84e3f30675"}).CronSend(c, &p)
}

// UserAddOrg 用户加入组织
func (s *DingSubscribe) UserAddOrg(c *gin.Context) {
	user := new(DingUser)
	userIdJson := s.EventJson["UserId"]
	userIdStrs := userIdJson.([]interface{})
	user.UserId = userIdStrs[0].(string)
	zap.L().Info("user.UserId: " + user.UserId)
	token, err2 := user.DingToken.GetAccessToken()
	if err2 != nil {
		zap.L().Error("user.DingToken.GetAccessToken() failed ,err: ", zap.Error(err2))
	} else {
		zap.L().Info("user.DingToken.GetAccessToken() success ,token: " + token)
	}
	user.DingToken.Token = token
	dingUser, _ := user.GetUserDetailByUserId()
	//输出用户信息，UserId和Token
	fmt.Printf("dingUser: %v\n", dingUser)
	//根据官方接口查询用户详细信息
	detailDingUser, err := dingUser.GetUserDetailByUserId()
	if err != nil {
		zap.L().Error("dingUser.GetUserByUserId() failed,err: ", zap.Error(err))
	}
	fmt.Printf("detailDingUser: %v\n", detailDingUser)

}

// UserLeaveOrg 用户退出组织
func (s *DingSubscribe) UserLeaveOrg(c *gin.Context) {
	user := new(DingUser)
	userIdJson := s.EventJson["UserId"]
	userIdStrs := userIdJson.([]interface{})
	user.UserId = userIdStrs[0].(string)
	zap.L().Info("user.UserId: " + user.UserId)
	token, err2 := user.DingToken.GetAccessToken()
	if err2 != nil {
		zap.L().Error("user.DingToken.GetAccessToken() failed ,err: ", zap.Error(err2))
	} else {
		zap.L().Info("user.DingToken.GetAccessToken() success ,token: " + token)
	}
	user.DingToken.Token = token
	dingUser, _ := user.GetUserDetailByUserId()
	//输出用户信息，UserId和Token
	fmt.Printf("dingUser: %v\n", dingUser)
	//根据官方接口查询用户详细信息
	detailDingUser, err := dingUser.GetUserDetailByUserId()
	if err != nil {
		zap.L().Error("dingUser.GetUserByUserId() failed,err: ", zap.Error(err))
	}
	fmt.Printf("detailDingUser: %v\n", detailDingUser)
}

// DingTalkCrypto 事件订阅加密
type DingTalkCrypto struct {
	Token          string
	EncodingAESKey string
	appkey         string
	BKey           []byte
	Block          cipher.Block
}

// NewDingTalkCrypto 新建钉钉密钥
func NewDingTalkCrypto(token, encodingAESKey, appkey string) *DingTalkCrypto {
	//fmt.Println(len(encodingAESKey))
	if len(encodingAESKey) != int(43) {
		panic("不合法的EncodingAESKey")
	}
	bkey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		panic(err.Error())
	}
	block, err := aes.NewCipher(bkey)
	if err != nil {
		panic(err.Error())
	}
	c := &DingTalkCrypto{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		appkey:         appkey,
		BKey:           bkey,
		Block:          block,
	}
	return c
}

// GetDecryptMsg 获取解密消息
func (c *DingTalkCrypto) GetDecryptMsg(signature, timestamp, nonce, secretMsg string) (string, error) {
	// 验证签名
	if !c.VerificationSignature(c.Token, timestamp, nonce, secretMsg, signature) {
		return "", errors.New("ERROR: 签名不匹配")
	}
	decode, err := base64.StdEncoding.DecodeString(secretMsg)
	if err != nil {
		return "", err
	}
	if len(decode) < aes.BlockSize {
		return "", errors.New("ERROR: 密文太短")
	}
	blockMode := cipher.NewCBCDecrypter(c.Block, c.BKey[:c.Block.BlockSize()])
	plantText := make([]byte, len(decode))
	blockMode.CryptBlocks(plantText, decode)
	plantText = pkCS7UnPadding(plantText)
	size := binary.BigEndian.Uint32(plantText[16:20])
	plantText = plantText[20:]
	corpID := plantText[size:]
	if string(corpID) != c.appkey {
		return "", errors.New("ERROR: CorpID匹配不正确")
	}
	return string(plantText[:size]), nil
}

// GetEncryptMsg 获取加密消息
func (c *DingTalkCrypto) GetEncryptMsg(msg string) (map[string]string, error) {
	// timestamp 获取时间戳
	var timestamp = time.Now().Second()
	// nonce 获取随机数
	var nonce = randomString(12)
	// str 加密msg; sign 签名(已加密); err:nil
	str, sign, err := c.GetEncryptMsgDetail(msg, fmt.Sprint(timestamp), nonce)
	//返回随机数，时间戳，加密字段，基于加密字段的签名
	return map[string]string{"nonce": nonce, "timeStamp": fmt.Sprint(timestamp), "encrypt": str, "msg_signature": sign}, err
}

// GetEncryptMsgDetail 获取加密详情
func (c *DingTalkCrypto) GetEncryptMsgDetail(msg, timestamp, nonce string) (string, string, error) {
	// 对传入的msg，timestamp，nonce加密开始
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	msg = randomString(16) + string(size) + msg + c.appkey
	plantText := pkCS7Padding([]byte(msg), c.Block.BlockSize())
	if len(plantText)%aes.BlockSize != 0 {
		return "", "", errors.New("ERROR: 消息体size不为16的倍数")
	}
	blockMode := cipher.NewCBCEncrypter(c.Block, c.BKey[:c.Block.BlockSize()])
	chipherText := make([]byte, len(plantText))
	blockMode.CryptBlocks(chipherText, plantText)
	outMsg := base64.StdEncoding.EncodeToString(chipherText)
	// 加密结束
	// signature 使用token，timestamp，nonce 以及加密的msg创建签名
	signature := c.CreateSignature(c.Token, timestamp, nonce, string(outMsg))
	// return 加密消息，签名，空
	return string(outMsg), signature, nil
}

func sha1Sign(s string) string {
	// The pattern for generating a hash is `sha1.New()`,
	// `sha1.Write(bytes)`, then `sha1.Sum([]byte{})`.
	// Here we start with a new hash.
	h := sha1.New()

	// `Write` expects bytes. If you have a string `s`,
	// use `[]byte(s)` to coerce it to bytes.
	h.Write([]byte(s))

	// This gets the finalized hash result as a byte
	// slice. The argument to `Sum` can be used to append
	// to an existing byte slice: it usually isn't needed.
	bs := h.Sum(nil)

	// SHA1 values are often printed in hex, for example
	// in git commits. Use the `%x` format verb to convert
	// a hash results to a hex string.
	return fmt.Sprintf("%x", bs)
}

// CreateSignature 数据签名
func (c *DingTalkCrypto) CreateSignature(token, timestamp, nonce, msg string) string {
	params := make([]string, 0)
	params = append(params, token)
	params = append(params, timestamp)
	params = append(params, nonce)
	params = append(params, msg)
	sort.Strings(params)
	return sha1Sign(strings.Join(params, ""))
}

// VerificationSignature 验证数据签名
func (c *DingTalkCrypto) VerificationSignature(token, timestamp, nonce, msg, sigture string) bool {
	return c.CreateSignature(token, timestamp, nonce, msg) == sigture
}

// pkCS7UnPadding 解密补位
func pkCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

// pkCS7Padding 加密补位
func pkCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// randomString 随机字符串
func randomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var randby bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randby = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[r.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[r.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return string(bytes)
}

// 测试
func main() {
	var ding = NewDingTalkCrypto("tokenxxxx", "o1w0aum42yaptlz8alnhwikjd3jenzt9cb9wmzptgus", "dingxxxxxx")
	msg, _ := ding.GetEncryptMsg("success")
	fmt.Println(msg)
	success, _ := ding.GetDecryptMsg("f36f4ba5337d426c7d4bca0dbcb06b3ddc1388fc", "1605695694141", "WelUQl6bCqcBa2fM", "X1VSe9cTJUMZu60d3kyLYTrBq5578ZRJtteU94wG0Q4Uk6E/wQYeJRIC0/UFW5Wkya1Ihz9oXAdLlyC9TRaqsQ==")
	fmt.Println(success)
}