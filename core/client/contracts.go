package client

import _ "embed"

//go:embed approval.teal
var ApproveTeal string

const ClearTeal = "" +
	"#pragma version 4\n\n" +
	"int 1"
