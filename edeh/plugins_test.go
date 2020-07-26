package main

import (
	"testing"
)

func TestBlacklits(t *testing.T) {
	pin := plugin{}
	if pin.Journal.Blacklist.Blacklisted("foo") {
		t.Error("a nil blacklist must not match any event")
	}
	pin.Journal.Blacklist = []string{}
	if !pin.Journal.Blacklist.Blacklisted("foo") {
		t.Error("a [] blacklist must match any event")
	}
	pin.Journal.Blacklist = []string{"bar"}
	if pin.Journal.Blacklist.Blacklisted("foo") {
		t.Error("blacklist [bar] must not match foo")
	}
	if !pin.Journal.Blacklist.Blacklisted("bar") {
		t.Error("blacklist [bar] must match bar")
	}
}
