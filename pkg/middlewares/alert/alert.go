package alert

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type LarkAlert struct {
	URL     string
	Secret  string
	Env     string
	Project string `json:"-"`
	Version string `json:"-"`

	SignFn func(int64) (string, error)
}

func (c *LarkAlert) IsZero() bool { return c == nil || c.URL == "" }

func (c *LarkAlert) Init() {
	if c.Secret != "" {
		c.SignFn = func(ts int64) (string, error) {
			payload := fmt.Sprintf("%v", ts) + "\n" + c.Secret

			var data []byte
			h := hmac.New(sha256.New, []byte(payload))
			_, err := h.Write(data)
			if err != nil {
				return "", err
			}

			signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
			return signature, nil
		}
	}
}

type ContentElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

type TitledContent struct {
	Title    string             `json:"title"`
	Contents [][]ContentElement `json:"content"`
}

type Content struct {
	Post map[string]TitledContent `json:"post"`
}

type Message struct {
	Timestamp int64   `json:"timestamp"`
	Sign      string  `json:"sign"`
	MsgType   string  `json:"msg_type"`
	Content   Content `json:"content"`
}

func (c *LarkAlert) Push(title, content string) error {
	req := &Message{
		Timestamp: time.Now().UTC().Unix(),
		Sign:      "",
		MsgType:   "post",
		Content: Content{
			Post: map[string]TitledContent{
				"en_ch": {
					Title: "WARNING [" + title + "]",
					Contents: [][]ContentElement{
						{{Tag: "text", Text: "env: " + c.Env}},
						{{Tag: "text", Text: "project: " + c.Project}},
						{{Tag: "text", Text: "version: " + c.Version}},
						{{Tag: "text", Text: content}},
					},
				},
			},
		},
	}

	if c.SignFn != nil {
		signature, err := c.SignFn(req.Timestamp)
		if err != nil {
			return err
		}
		req.Sign = signature
	}

	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(req)

	rsp, err := http.Post(c.URL, "application/json", buf)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if len(body) == 0 {
		return nil
	}

	v := &struct {
		Code int    `json:"cod"`
		Msg  string `json:"msg"`
	}{}
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}
	switch v.Code {
	case 19021:
		return errors.New("invalid message")
	case 9499:
		return errors.New("invalid signature")
	case 0:
		return nil
	default:
		return errors.Errorf("code %d message %s", v.Code, v.Msg)
	}
}
