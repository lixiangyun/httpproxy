package main

import (
	"log"
	"regexp"
	"strings"
)

type Match interface {
	Do(value string) bool
}

type MatchRegex struct {
	exp string
	reg *regexp.Regexp
}

func NewMatchRegex(exp string) Match {
	reg, err := regexp.Compile(exp)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return &MatchRegex{exp: exp, reg: reg}
}

func (m *MatchRegex) Do(value string) bool {
	return m.reg.MatchString(value)
}

type MatchPrefix struct {
	exp string
}

func NewMatchPrefix(exp string) Match {
	return &MatchPrefix{exp: exp}
}

func (m *MatchPrefix) Do(value string) bool {
	if 0 == strings.Index(value, m.exp) {
		return true
	}
	return false
}

type MatchFull struct {
	exp string
}

func NewMatchFull(exp string) Match {
	return &MatchFull{exp: exp}
}

func (m *MatchFull) Do(value string) bool {
	if 0 == strings.Compare(m.exp, value) {
		return true
	}
	return false
}

func NewMatch(t MatchType, v string) Match {
	switch t {
	case "prefix":
		{
			return NewMatchPrefix(v)
		}
	case "regex":
		{
			return NewMatchRegex(v)
		}
	case "equal":
		{
			return NewMatchFull(v)
		}
	}
	return nil
}
