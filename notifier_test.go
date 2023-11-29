package main

import "testing"

func Test_parseTags(t *testing.T) {
	s := "zigzigcheers/todo:main\nzigzigcheers/todo:sha-82e4bb3"
	tag, e := parseTags([]string{s})
	if e != nil {
		t.Error(e)
		return
	}
	if tag != `sha-82e4bb3` {
		t.Error("tag is not `sha-` , but ", tag)
		return
	}
}
