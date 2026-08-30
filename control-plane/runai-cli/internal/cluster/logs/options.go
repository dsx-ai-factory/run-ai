package logs

import (
	"errors"
	"net/url"
	"strconv"
	"time"
)

const (
	queryParamPod                   = "pod"
	queryParamContainer             = "container"
	queryParamTailLines             = "tailLines"
	queryParamTimestamps            = "timestamps"
	queryParamFollow                = "follow"
	queryParamPrevious              = "previous"
	queryParamLimitBytes            = "limitBytes"
	queryParamSinceSeconds          = "sinceSeconds"
	queryParamSinceTime             = "sinceTime"
	queryParamTrue                  = "true"
	ErrSinceTimeAndSecondsSpecified = "at most one of `sinceTime` or `sinceSeconds` may be specified"
	ErrLimitBytesNegative           = "limit-bytes must be greater than 0"
	ErrSinceSecondsNegative         = "sinceSeconds must be greater than 0"
	ErrTailLinesInvalid             = "tail must be greater than or equal to -1"
)

type Options struct {
	Pod        string
	Container  string
	TailLines  int64
	Timestamps bool
	Follow     bool
	Previous   bool
	LimitBytes int64
	Since      time.Duration
	SinceTime  string
}

func NewOptions(pod, container, sinceTime string,
	tailLines int64, timestamps, follow, previous bool,
	limitBytes int64, since time.Duration) *Options {
	return &Options{
		Pod:        pod,
		Container:  container,
		TailLines:  tailLines,
		Timestamps: timestamps,
		Follow:     follow,
		Previous:   previous,
		LimitBytes: limitBytes,
		Since:      since,
		SinceTime:  sinceTime,
	}
}

func (o Options) Validate() error {
	if len(o.SinceTime) > 0 && o.Since != 0 {
		return errors.New(ErrSinceTimeAndSecondsSpecified)
	}

	if o.LimitBytes < 0 {
		return errors.New(ErrLimitBytesNegative)
	}

	if o.Since < 0 {
		return errors.New(ErrSinceSecondsNegative)
	}

	if o.TailLines < -1 {
		return errors.New(ErrTailLinesInvalid)
	}

	return nil
}

func (o Options) ToQueryParams() (*url.Values, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	vars := url.Values{}
	if o.Pod != "" {
		vars.Set(queryParamPod, o.Pod)
	}
	if o.Container != "" {
		vars.Set(queryParamContainer, o.Container)
	}
	if o.TailLines != -1 {
		// convert int64 to string
		vars.Set(queryParamTailLines, strconv.FormatInt(o.TailLines, 10))
	}
	if o.Timestamps {
		vars.Set(queryParamTimestamps, queryParamTrue)
	}
	if o.Follow {
		vars.Set(queryParamFollow, queryParamTrue)
	}
	if o.Previous {
		vars.Set(queryParamPrevious, queryParamTrue)
	}
	if o.LimitBytes > 0 {
		vars.Set(queryParamLimitBytes, strconv.FormatInt(o.LimitBytes, 10))
	}
	if o.Since > 0 {
		vars.Set(queryParamSinceSeconds, strconv.FormatInt(int64(o.Since.Seconds()), 10))
	}
	if o.SinceTime != "" {
		vars.Set(queryParamSinceTime, o.SinceTime)
	}
	return &vars, nil
}
