package svc

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	_ yaml.BytesMarshaler   = RemoteConfig_UNKNOWN
	_ yaml.BytesUnmarshaler = (*RemoteConfig_RemoteType)(nil)
)

func (t RemoteConfig_RemoteType) MarshalYAML() ([]byte, error) {
	name, valid := RemoteConfig_RemoteType_name[int32(t)]
	if !valid {
		return yaml.Marshal(int32(t))
	}

	return yaml.Marshal(name)
}

func (t *RemoteConfig_RemoteType) UnmarshalYAML(data []byte) error {
	var i int32
	err := yaml.Unmarshal(data, &i)
	// err is nil, the value is successfully parsed.
	if err == nil {
		_, valid := RemoteConfig_RemoteType_name[i]
		if !valid {
			return fmt.Errorf("integer value %d is not a valid RemoteType", i)
		}

		*t = RemoteConfig_RemoteType(i)
		return nil
	}

	var s string
	err = yaml.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal as string or int: %w", err)
	}

	uppedS := strings.ToUpper(s)
	v, valid := RemoteConfig_RemoteType_value[uppedS]
	if !valid {
		return fmt.Errorf("string %s is not a valid remote type", s)
	}

	*t = RemoteConfig_RemoteType(v)
	return nil
}

func (c *GiTrimConfig) GetProperShutdownWaitSecs() int {
	if c.ShutdownWaitSecs == 0 {
		return 90
	}

	return int(c.ShutdownWaitSecs)
}
