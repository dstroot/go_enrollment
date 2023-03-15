package client

import (
  "errors"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "reflect"
  "testing"
)

func TestWhtever(t *testing.T) {
  tests := []struct {
    endpoint url.URL
    prefix   string
    key      string
    want     url.URL
  }{
    // key is empty, no problem
    {
      endpoint: url.URL{Scheme: "http", Host: "example.com", Path: "/v2/keys"},
      prefix:   "",
      key:      "",
      want:     url.URL{Scheme: "http", Host: "example.com", Path: "/v2/keys"},
    },

    // key is joined to path
    {
      endpoint: url.URL{Scheme: "http", Host: "example.com", Path: "/v2/keys"},
      prefix:   "",
      key:      "/foo/bar",
      want:     url.URL{Scheme: "http", Host: "example.com", Path: "/v2/keys/foo/bar"},
    },

    // key is joined to path when path is empty
    {
      endpoint: url.URL{Scheme: "http", Host: "example.com", Path: ""},
      prefix:   "",
      key:      "/foo/bar",
      want:     url.URL{Scheme: "http", Host: "example.com", Path: "/foo/bar"},
    },

    // Host field carries through with port
    {
      endpoint: url.URL{Scheme: "http", Host: "example.com:8080", Path: "/v2/keys"},
      prefix:   "",
      key:      "",
      want:     url.URL{Scheme: "http", Host: "example.com:8080", Path: "/v2/keys"},
    },

    // Scheme carries through
    {
      endpoint: url.URL{Scheme: "https", Host: "example.com", Path: "/v2/keys"},
      prefix:   "",
      key:      "",
      want:     url.URL{Scheme: "https", Host: "example.com", Path: "/v2/keys"},
    },
    // Prefix is applied
    {
      endpoint: url.URL{Scheme: "https", Host: "example.com", Path: "/foo"},
      prefix:   "/bar",
      key:      "/baz",
      want:     url.URL{Scheme: "https", Host: "example.com", Path: "/foo/bar/baz"},
    },
  }

  for i, tt := range tests {
    got := v2KeysURL(tt.endpoint, tt.prefix, tt.key)
    if tt.want != *got {
      t.Errorf("#%d: want=%#v, got=%#v", i, tt.want, *got)
    }
  }
}
