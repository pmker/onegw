/*
 * Copyright 2019 The go-deposit Authors
 * This file is part of the go-deposit library.
 *
 * The go-deposit library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The go-deposit library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the go-deposit library. If not, see <http://www.gnu.org/licenses/>.
 */

package Queue

// element in list
type element struct {
	key  string
	next *element
}

type UniqueList interface {
	Append(key string, data interface{})
	Shift() (key string, data interface{})
	UnShift(key string, data interface{})
	Traverse(handler func(key string, data interface{}) bool)
	Filter(filter func(key string, data interface{}) bool)
	Size() int
	Clear()
}

type Queue struct {
	m    map[string]interface{}
	head *element
	tail *element
}

// New create a linked list
func New() UniqueList {
	head := &element{}
	return &Queue{
		m:    make(map[string]interface{}),
		head: head,
		tail: head,
	}
}

// Append an element at tail of list
func (l *Queue) Append(key string, data interface{}) {
	if _, ok := l.m[key]; ok {
		return
	}

	l.m[key] = data
	e := &element{
		key: key,
	}

	l.tail.next = e
	l.tail = e
}

// Shift will remove the first element of list
func (l *Queue) Shift() (key string, data interface{}) {
	e := l.head.next
	if e == nil {
		return
	}

	v := l.m[e.key]

	l.remove(l.head, e)

	return e.key, v
}

// UnShift use to insert an element to front of list
func (l *Queue) UnShift(key string, data interface{}) {
	if _, ok := l.m[key]; ok {
		return
	}

	l.m[key] = data

	e := &element{
		key: key,
	}

	e.next = l.head.next
	l.head.next = e
}

// Remove remove an element from list
func (l *Queue) remove(prev, current *element) {
	prev.next = current.next
	if current.next == nil {
		l.tail = prev
	}

	delete(l.m, current.key)
}

// Traverse list, if handler return false, will not iterate forward elements
func (l *Queue) Traverse(handler func(key string, value interface{}) bool) {
	for current := l.head.next; current != nil; current = current.next {
		if !handler(current.key, l.m[current.key]) {
			break
		}
	}
}

// Filter, if filter return true, will remove those elements
func (l *Queue) Filter(filter func(key string, value interface{}) bool) {
	for prev, current := l.head, l.head.next; current != nil; {
		if filter(current.key, l.m[current.key]) {
			// remove
			l.remove(prev, current)
			current = prev.next
		} else {
			prev, current = current, current.next
		}
	}
}

// Size indicate how many elements in list
func (l *Queue) Size() int {
	return len(l.m)
}

// Clear remove all elements from list
func (l *Queue) Clear() {
	l.m = make(map[string]interface{})
	l.head.next = nil
	l.tail = l.head
}
