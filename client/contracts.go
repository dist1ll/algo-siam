package client

import _ "embed"

//go:embed approval.teal
var ApproveTeal string

//go:embed clear.teal
var ClearTeal string

// Schema of AlgorandBuffer.

const LocalInts = 0
const LocalBytes = 0
const GlobalInts = 0
const GlobalBytes = 64
