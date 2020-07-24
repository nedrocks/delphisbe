// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import twitter "github.com/dghubble/go-twitter/twitter"

// TwitterClient is an autogenerated mock type for the TwitterClient type
type TwitterClient struct {
	mock.Mock
}

// LookupUsers provides a mock function with given fields: screenNames
func (_m *TwitterClient) LookupUsers(screenNames []string) ([]twitter.User, error) {
	ret := _m.Called(screenNames)

	var r0 []twitter.User
	if rf, ok := ret.Get(0).(func([]string) []twitter.User); ok {
		r0 = rf(screenNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]twitter.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(screenNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchUsers provides a mock function with given fields: query, page, count
func (_m *TwitterClient) SearchUsers(query string, page int, count int) ([]twitter.User, error) {
	ret := _m.Called(query, page, count)

	var r0 []twitter.User
	if rf, ok := ret.Get(0).(func(string, int, int) []twitter.User); ok {
		r0 = rf(query, page, count)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]twitter.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(query, page, count)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}