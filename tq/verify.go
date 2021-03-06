package tq

import (
	"net/http"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const (
	maxVerifiesConfigKey     = "lfs.transfer.maxverifies"
	defaultMaxVerifyAttempts = 3
)

func verifyUpload(c *lfsapi.Client, t *Transfer) error {
	action, err := t.Actions.Get("verify")
	if err != nil {
		if IsActionMissingError(err) {
			return nil
		}
		return err
	}

	req, err := http.NewRequest("POST", action.Href, nil)
	if err != nil {
		return err
	}

	err = lfsapi.MarshalToRequest(req, struct {
		Oid  string `json:"oid"`
		Size int64  `json:"size"`
	}{Oid: t.Oid, Size: t.Size})
	if err != nil {
		return err
	}

	for key, value := range action.Header {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/vnd.git-lfs+json")

	mv := c.GitEnv().Int(maxVerifiesConfigKey, defaultMaxVerifyAttempts)
	mv = tools.MaxInt(defaultMaxVerifyAttempts, mv)

	for i := 1; i <= mv; i++ {
		tracerx.Printf("tq: verify %s attempt #%d (max: %d)", t.Oid[:7], i, mv)

		var res *http.Response

		if res, err = c.Do(req); err != nil {
			tracerx.Printf("tq: verify err: %+v", err.Error())
		} else {
			err = res.Body.Close()
			break
		}
	}
	return err
}
