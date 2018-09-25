package main

import (
	"github.com/kelseyhightower/envconfig"
)

// Specification are env variables used by ems
// split_words is used by envconfig
type Specification struct {
	EtcdEndpoint           string `split_words:"true"`
	EtcdPrefix             string `split_words:"true"`
	B2ApplicationID        string `split_words:"true"`
	B2ApplicationKey       string `split_words:"true"`
	B2BucketID             string `split_words:"true"`
	B2Object               string `split_words:"true"`
	B2UploadRetryInterval  int    `split_words:"true"`
	WaitForChangesInterval int    `split_words:"true"`
}

// ParseENV pareses env for variables
func ParseENV() (*Specification, error) {
	s := Specification{}

	err := envconfig.Process("ESB", &s)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
