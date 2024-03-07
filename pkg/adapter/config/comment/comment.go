// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package comment provides facilities for parsing comments out of a
// YAML sequence or mapping node recursively, keeping them as a Comment
// struct instance, and merging them in another YAML parsed node, so
// comments may be preserved when that YAML node is serialized again.
package comment

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Comment struct contains the head-comments for a sequence or mapping
// yaml node instance.
type Comment struct {
	m *Map // comments for a mapping yaml node (or nil otherwise)
	s *Seq // comments for a sequence yaml node (or nil otherwise)
}

// Map struct contains head-comments which are written right before
// each mapping key element. If those keys are mapped to some other
// sequence or mapping yaml nodes themselves, they may be parsed as
// nested Comment instances. This Map struct also contains a mapping
// from the key elements to their nested Comment instances (if any).
type Map struct {
	keyComments map[string]string   // key name -> its head comment
	mapComments map[string]*Comment // key name -> its inner comments
}

// Seq struct contains head-comments which are written right before
// each sequence element. If those elements contain some other sequence
// or mapping yaml nodes themselves, they may be parsed as nested
// Comment instances. This Seq struct also contains a slice of *Comment
// instances which indicates relevant nested comments. If some of the
// sequence members need a nested Comment instance and some of them do
// not need it, the slice of *Comment instances can contain nil elements
// jumping over the missing *Comment instances.
type Seq struct {
	topComments []string   // comments on top-level items
	seqComments []*Comment // sequence of nil or nested comments
}

// LoadFrom expects a yaml node with the sequence or mapping kind,
// iterates over its keys, and loads their HeadComment strings in the
// created Comment instance. If contained values had sequence or mapping
// kinds themselves, they will be loaded recursively into new Comment
// instances and kept in returned Comment, so they can be saved again.
func LoadFrom(n *yaml.Node) (*Comment, error) {
	switch n.Kind {
	case yaml.SequenceNode:
		return loadSeq(n)
	case yaml.MappingNode:
		return loadMap(n)
	default:
		return nil, errors.New("node must be a mapping or a sequence")
	}
}

func loadSeq(n *yaml.Node) (*Comment, error) {
	c := &Comment{
		s: &Seq{},
	}
	for i, cn := range n.Content {
		c.s.topComments = append(c.s.topComments, cn.HeadComment)
		switch cn.Kind {
		case yaml.SequenceNode, yaml.MappingNode:
			nested, err := LoadFrom(cn)
			if err != nil {
				return nil, fmt.Errorf(
					"loading nested comments for i=%d: %w", i, err,
				)
			}
			c.s.seqComments = append(c.s.seqComments, nested)
		default:
			c.s.seqComments = append(c.s.seqComments, nil)
		}
	}
	return c, nil
}

func loadMap(n *yaml.Node) (*Comment, error) {
	c := &Comment{
		m: &Map{
			keyComments: make(map[string]string),
			mapComments: make(map[string]*Comment),
		},
	}
	var key string
	for i, cn := range n.Content {
		switch i % 2 {
		case 0:
			key = cn.Value
			c.m.keyComments[key] = cn.HeadComment
		case 1:
			switch cn.Kind {
			case yaml.SequenceNode, yaml.MappingNode:
				nested, err := LoadFrom(cn)
				if err != nil {
					return nil, fmt.Errorf(
						"loading nested comments from %q key: %w",
						key, err,
					)
				}
				c.m.mapComments[key] = nested
			}
		}
	}
	return c, nil
}

// SaveInto saves comments which are recorded in the `c` instance
// into the given `n` yaml node. The `n` argument is expected to have
// a sequence or mapping kind. If its contained values had sequence or
// mapping kinds themselves, they will be checked recursively too (if
// they had some corresponding comments among the `c` nested comments).
func (c *Comment) SaveInto(n *yaml.Node) error {
	if c == nil {
		return nil // there is no comments to be saved
	}
	switch k := n.Kind; k {
	case yaml.SequenceNode:
		if c.s == nil {
			return errors.New("unexpected sequence node")
		}
		if err := c.s.saveInto(n); err != nil {
			return fmt.Errorf("saving a sequence comments: %w", err)
		}
	case yaml.MappingNode:
		if c.m == nil {
			return errors.New("unexpected mapping node")
		}
		if err := c.m.saveInto(n); err != nil {
			return fmt.Errorf("saving a mapping comments: %w", err)
		}
	default:
		return fmt.Errorf("expected a mapping or sequence (kind=%d)", k)
	}
	return nil
}

func (s *Seq) saveInto(n *yaml.Node) error {
	for i, cn := range n.Content {
		if i >= len(s.topComments) {
			return nil
		}
		cn.HeadComment = s.topComments[i]
		nested := s.seqComments[i]
		if nested == nil {
			continue
		}
		if err := nested.SaveInto(cn); err != nil {
			return fmt.Errorf(
				"saving nested comments for i=%d: %w", i, err,
			)
		}
	}
	return nil
}

func (m *Map) saveInto(n *yaml.Node) error {
	var key string
	for i, cn := range n.Content {
		switch i % 2 {
		case 0:
			key = cn.Value
			cn.HeadComment = m.keyComments[key]
		case 1:
			nested, exists := m.mapComments[key]
			if !exists {
				continue
			}
			if err := nested.SaveInto(cn); err != nil {
				return fmt.Errorf(
					"saving nested comments into %q key: %w", key, err,
				)
			}
		}
	}
	return nil
}
