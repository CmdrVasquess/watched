//go:build !windows

package watched

func FindJournalDir() (dir string, err error) { return ".", nil }
