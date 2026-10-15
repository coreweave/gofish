//
// SPDX-License-Identifier: BSD-3-Clause
//

package common

import "errors"

var (
	ErrIsEmpty    = errors.New("value is empty")
	ErrIsNegative = errors.New("value is negative")

	// ErrArrayMemberUpdate is returned by Update on a member of an array in
	// its parent resource, whose @odata.id ends in a fragment such as
	// "#/Fans/0". The service ignores the fragment, so a PATCH there would
	// reach the parent with the member's properties at the top level.
	ErrArrayMemberUpdate = errors.New("cannot update an array member at its own URI")
)
