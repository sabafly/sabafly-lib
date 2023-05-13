/*
	Copyright (C) 2022-2023  sabafly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package permissions

import (
	"strings"
)

func New() Permission {
	return make(map[string]Permission)
}

type Permission map[string]Permission

func (p Permission) Has(perm string) bool {
	perms := strings.Split(perm, ".")
	p2 := p
	for _, v := range perms {
		if _, ok := p2["*"]; ok {
			return true
		}
		var ok bool
		p2, ok = p2[v]
		if !ok && v != "*" {
			return false
		}
	}
	return true
}

func (p Permission) Add(perm string) {
	perms := strings.Split(perm, ".")
	p2 := p
	for _, v := range perms {
		if _, ok := p2[v]; !ok {
			p2[v] = New()
		}
		p2 = p2[v]
	}
}

func (p Permission) Del(perm string) {
	perms := strings.Split(perm, ".")
	p2 := p
	for i, v := range perms {
		if _, ok := p2[v]; !ok {
			return
		}
		if i >= len(perms)-1 {
			delete(p2, v)
		}
		p2 = p2[v]
	}
}
