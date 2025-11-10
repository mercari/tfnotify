package mask

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func ParseMasksFromEnv() ([]*config.Mask, error) {
	return ParseMasks(os.Getenv("TFNOTIFY_MASKS"), os.Getenv("TFNOTIFY_MASKS_SEPARATOR"))
}

func ParseMasks(maskStr, maskSep string) ([]*config.Mask, error) {
	if maskStr == "" {
		return nil, nil
	}
	if maskSep == "" {
		maskSep = "," // default separator
	}
	maskStrs := strings.Split(maskStr, maskSep)
	masks := make([]*config.Mask, 0, len(maskStrs))
	for _, maskStr := range maskStrs {
		mask, err := parseMask(maskStr)
		if err != nil {
			return nil, fmt.Errorf("parse a mask: %w", logerr.WithFields(err, logrus.Fields{
				"mask": maskStr,
			}))
		}
		if mask == nil {
			continue
		}
		masks = append(masks, mask)
	}
	return masks, nil
}

func parseMask(maskStr string) (*config.Mask, error) {
	typ, value, ok := strings.Cut(maskStr, ":")
	if !ok {
		return nil, errors.New("the mask is invalid. ':' is missing")
	}
	switch typ {
	case "env":
		if e := os.Getenv(value); e != "" {
			return &config.Mask{
				Type:  "equal",
				Value: e,
			}, nil
		}
		// the environment variable is missing
		return nil, nil //nolint:nilnil
	case "regexp":
		p, err := regexp.Compile(value)
		if err != nil {
			return nil, fmt.Errorf("the regular expression is invalid: %w", err)
		}
		return &config.Mask{
			Type:   "regexp",
			Value:  value,
			Regexp: p,
		}, nil
	default:
		return nil, errors.New("the mask type is invalid")
	}
}
