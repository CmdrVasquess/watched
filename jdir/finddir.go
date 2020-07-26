// +build !windows

package jdir

func FindJournalDir() (dir string, err error) { return ".", nil }
