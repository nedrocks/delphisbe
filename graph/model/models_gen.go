// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type DiscussionNotificationPreferences interface {
	IsDiscussionNotificationPreferences()
}

type Entity interface {
	IsEntity()
}

type AddDiscussionParticipantInput struct {
	GradientColor *GradientColor `json:"gradientColor"`
	FlairID       *string        `json:"flairID"`
	HasJoined     *bool          `json:"hasJoined"`
	IsAnonymous   bool           `json:"isAnonymous"`
}

type ConciergeContent struct {
	AppActionID *string            `json:"appActionID"`
	MutationID  *string            `json:"mutationID"`
	Options     []*ConciergeOption `json:"options"`
}

type ConciergeOption struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected"`
}

type DiscussionInput struct {
	AnonymityType *AnonymityType `json:"anonymityType"`
	Title         *string        `json:"title"`
	AutoPost      *bool          `json:"autoPost"`
	IdleMinutes   *int           `json:"idleMinutes"`
	PublicAccess  *bool          `json:"publicAccess"`
	IconURL       *string        `json:"iconURL"`
}

type Media struct {
	ID                string             `json:"id"`
	CreatedAt         string             `json:"createdAt"`
	IsDeleted         bool               `json:"isDeleted"`
	DeletedReasonCode *PostDeletedReason `json:"deletedReasonCode"`
	MediaType         *string            `json:"mediaType"`
	MediaSize         *MediaSize         `json:"mediaSize"`
	AssetLocation     *string            `json:"assetLocation"`
}

type MediaSize struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	SizeKb float64 `json:"sizeKb"`
}

type ParticipantNotificationPreferences struct {
	ID string `json:"id"`
}

func (ParticipantNotificationPreferences) IsDiscussionNotificationPreferences() {}

type ParticipantProfile struct {
	IsAnonymous   *bool          `json:"isAnonymous"`
	Flair         *Flair         `json:"flair"`
	GradientColor *GradientColor `json:"gradientColor"`
}

type PollInput struct {
	PollText string    `json:"pollText"`
	EndTime  time.Time `json:"endTime"`
	Option1  string    `json:"option1"`
	Option2  string    `json:"option2"`
	Option3  *string   `json:"option3"`
	Option4  *string   `json:"option4"`
}

type PostContentInput struct {
	PostText          string     `json:"postText"`
	PostType          PostType   `json:"postType"`
	MentionedEntities []string   `json:"mentionedEntities"`
	QuotedPostID      *string    `json:"quotedPostID"`
	MediaID           *string    `json:"mediaID"`
	Poll              *PollInput `json:"poll"`
	ImportedContentID *string    `json:"importedContentID"`
	Preview           *string    `json:"preview"`
}

type URL struct {
	DisplayText string `json:"displayText"`
	URL         string `json:"url"`
}

type UnknownEntity struct {
	ID string `json:"id"`
}

func (UnknownEntity) IsEntity() {}

type UpdateParticipantInput struct {
	GradientColor   *GradientColor `json:"gradientColor"`
	IsUnsetGradient *bool          `json:"isUnsetGradient"`
	FlairID         *string        `json:"flairID"`
	IsUnsetFlairID  *bool          `json:"isUnsetFlairID"`
	IsAnonymous     *bool          `json:"isAnonymous"`
	HasJoined       *bool          `json:"hasJoined"`
}

type ViewerNotificationPreferences struct {
	ID string `json:"id"`
}

func (ViewerNotificationPreferences) IsDiscussionNotificationPreferences() {}

type AnonymityType string

const (
	AnonymityTypeUnknown AnonymityType = "UNKNOWN"
	AnonymityTypeWeak    AnonymityType = "WEAK"
	AnonymityTypeStrong  AnonymityType = "STRONG"
)

var AllAnonymityType = []AnonymityType{
	AnonymityTypeUnknown,
	AnonymityTypeWeak,
	AnonymityTypeStrong,
}

func (e AnonymityType) IsValid() bool {
	switch e {
	case AnonymityTypeUnknown, AnonymityTypeWeak, AnonymityTypeStrong:
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

type GradientColor string

const (
	GradientColorUnknown    GradientColor = "UNKNOWN"
	GradientColorMauve      GradientColor = "MAUVE"
	GradientColorFuschia    GradientColor = "FUSCHIA"
	GradientColorCinnabar   GradientColor = "CINNABAR"
	GradientColorVermillion GradientColor = "VERMILLION"
	GradientColorCerulean   GradientColor = "CERULEAN"
	GradientColorTurquoise  GradientColor = "TURQUOISE"
	GradientColorCeladon    GradientColor = "CELADON"
	GradientColorTaupe      GradientColor = "TAUPE"
	GradientColorSaffron    GradientColor = "SAFFRON"
	GradientColorViridian   GradientColor = "VIRIDIAN"
	GradientColorChartruese GradientColor = "CHARTRUESE"
	GradientColorLavender   GradientColor = "LAVENDER"
	GradientColorGoldenrod  GradientColor = "GOLDENROD"
	GradientColorSeafoam    GradientColor = "SEAFOAM"
	GradientColorAzalea     GradientColor = "AZALEA"
	GradientColorViolet     GradientColor = "VIOLET"
	GradientColorMahogany   GradientColor = "MAHOGANY"
)

var AllGradientColor = []GradientColor{
	GradientColorUnknown,
	GradientColorMauve,
	GradientColorFuschia,
	GradientColorCinnabar,
	GradientColorVermillion,
	GradientColorCerulean,
	GradientColorTurquoise,
	GradientColorCeladon,
	GradientColorTaupe,
	GradientColorSaffron,
	GradientColorViridian,
	GradientColorChartruese,
	GradientColorLavender,
	GradientColorGoldenrod,
	GradientColorSeafoam,
	GradientColorAzalea,
	GradientColorViolet,
	GradientColorMahogany,
}

func (e GradientColor) IsValid() bool {
	switch e {
	case GradientColorUnknown, GradientColorMauve, GradientColorFuschia, GradientColorCinnabar, GradientColorVermillion, GradientColorCerulean, GradientColorTurquoise, GradientColorCeladon, GradientColorTaupe, GradientColorSaffron, GradientColorViridian, GradientColorChartruese, GradientColorLavender, GradientColorGoldenrod, GradientColorSeafoam, GradientColorAzalea, GradientColorViolet, GradientColorMahogany:
		return true
	}
	return false
}

func (e GradientColor) String() string {
	return string(e)
}

func (e *GradientColor) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = GradientColor(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid GradientColor", str)
	}
	return nil
}

func (e GradientColor) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type InviteRequestStatus string

const (
	InviteRequestStatusAccepted  InviteRequestStatus = "ACCEPTED"
	InviteRequestStatusRejected  InviteRequestStatus = "REJECTED"
	InviteRequestStatusPending   InviteRequestStatus = "PENDING"
	InviteRequestStatusCancelled InviteRequestStatus = "CANCELLED"
)

var AllInviteRequestStatus = []InviteRequestStatus{
	InviteRequestStatusAccepted,
	InviteRequestStatusRejected,
	InviteRequestStatusPending,
	InviteRequestStatusCancelled,
}

func (e InviteRequestStatus) IsValid() bool {
	switch e {
	case InviteRequestStatusAccepted, InviteRequestStatusRejected, InviteRequestStatusPending, InviteRequestStatusCancelled:
		return true
	}
	return false
}

func (e InviteRequestStatus) String() string {
	return string(e)
}

func (e *InviteRequestStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = InviteRequestStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid InviteRequestStatus", str)
	}
	return nil
}

func (e InviteRequestStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Platform string

const (
	PlatformUnknown Platform = "UNKNOWN"
	PlatformIos     Platform = "IOS"
	PlatformAndroid Platform = "ANDROID"
	PlatformWeb     Platform = "WEB"
)

var AllPlatform = []Platform{
	PlatformUnknown,
	PlatformIos,
	PlatformAndroid,
	PlatformWeb,
}

func (e Platform) IsValid() bool {
	switch e {
	case PlatformUnknown, PlatformIos, PlatformAndroid, PlatformWeb:
		return true
	}
	return false
}

func (e Platform) String() string {
	return string(e)
}

func (e *Platform) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Platform(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Platform", str)
	}
	return nil
}

func (e Platform) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostDeletedReason string

const (
	PostDeletedReasonUnknown            PostDeletedReason = "UNKNOWN"
	PostDeletedReasonModeratorRemoved   PostDeletedReason = "MODERATOR_REMOVED"
	PostDeletedReasonParticipantRemoved PostDeletedReason = "PARTICIPANT_REMOVED"
)

var AllPostDeletedReason = []PostDeletedReason{
	PostDeletedReasonUnknown,
	PostDeletedReasonModeratorRemoved,
	PostDeletedReasonParticipantRemoved,
}

func (e PostDeletedReason) IsValid() bool {
	switch e {
	case PostDeletedReasonUnknown, PostDeletedReasonModeratorRemoved, PostDeletedReasonParticipantRemoved:
		return true
	}
	return false
}

func (e PostDeletedReason) String() string {
	return string(e)
}

func (e *PostDeletedReason) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostDeletedReason(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostDeletedReason", str)
	}
	return nil
}

func (e PostDeletedReason) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostType string

const (
	PostTypeStandard        PostType = "STANDARD"
	PostTypeImportedContent PostType = "IMPORTED_CONTENT"
	PostTypeAlert           PostType = "ALERT"
	PostTypeConcierge       PostType = "CONCIERGE"
)

var AllPostType = []PostType{
	PostTypeStandard,
	PostTypeImportedContent,
	PostTypeAlert,
	PostTypeConcierge,
}

func (e PostType) IsValid() bool {
	switch e {
	case PostTypeStandard, PostTypeImportedContent, PostTypeAlert, PostTypeConcierge:
		return true
	}
	return false
}

func (e PostType) String() string {
	return string(e)
}

func (e *PostType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostType", str)
	}
	return nil
}

func (e PostType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
