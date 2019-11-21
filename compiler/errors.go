package compiler

// This file contains some utility functions related to error handling.

import (
	"go/token"
	"go/types"
)

// makeError makes it easy to create an error from a token.Pos with a message.
func (c *compilerContext) makeError(pos token.Pos, msg string) types.Error {
	return types.Error{
		Fset: c.ir.Program.Fset,
		Pos:  pos,
		Msg:  msg,
	}
}

func (c *Compiler) addError(pos token.Pos, msg string) {
	c.diagnostics = append(c.diagnostics, c.makeError(pos, msg))
}
