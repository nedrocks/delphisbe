// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type DiscussionNotificationPreferences interface {
	IsDiscussionNotificationPreferences()
}

type ParticipantNotificationPreferences struct {
	ID string `json:"id"`
}

func (ParticipantNotificationPreferences) IsDiscussionNotificationPreferences() {}

type ViewerNotificationPreferences struct {
	ID string `json:"id"`
}

func (ViewerNotificationPreferences) IsDiscussionNotificationPreferences() {}

type AnonymityType string

const (
	AnonymityTypeWeak   AnonymityType = "WEAK"
	AnonymityTypeStrong AnonymityType = "STRONG"
)

var AllAnonymityType = []AnonymityType{
	AnonymityTypeWeak,
	AnonymityTypeStrong,
}

func (e AnonymityType) IsValid() bool {
	switch e {
	case AnonymityTypeWeak, AnonymityTypeStrong:
		return true
	}
	return false
}

func (e AnonymityType) String() string {
	return string(e)
}

func (e *AnonymityType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AnonymityType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AnonymityType", str)
	}
	return nil
}

func (e AnonymityType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
