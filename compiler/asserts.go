package compiler

// This file implements functions that do certain safety checks that are
// required by the Go programming language.

import (
	"go/types"

	"tinygo.org/x/go-llvm"
)

// createLookupBoundsCheck emits a bounds check before doing a lookup into a
// slice. This is required by the Go language spec: an index out of bounds must
// cause a panic.
func (b *builder) createLookupBoundsCheck(arrayLen, index llvm.Value, indexType types.Type) {
	if b.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	if index.Type().IntTypeWidth() < arrayLen.Type().IntTypeWidth() {
		// Sometimes, the index can be e.g. an uint8 or int8, and we have to
		// correctly extend that type.
		if indexType.Underlying().(*types.Basic).Info()&types.IsUnsigned == 0 {
			index = b.CreateZExt(index, arrayLen.Type(), "")
		} else {
			index = b.CreateSExt(index, arrayLen.Type(), "")
		}
	} else if index.Type().IntTypeWidth() > arrayLen.Type().IntTypeWidth() {
		// The index is bigger than the array length type, so extend it.
		arrayLen = b.CreateZExt(arrayLen, index.Type(), "")
	}

	faultBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, "lookup.outofbounds")
	nextBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, "lookup.next")
	b.blockExits[b.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Now do the bounds check: index >= arrayLen
	outOfBounds := b.CreateICmp(llvm.IntUGE, index, arrayLen, "")
	b.CreateCondBr(outOfBounds, faultBlock, nextBlock)

	// Fail: this is a nil pointer, exit with a panic.
	b.SetInsertPointAtEnd(faultBlock)
	b.createRuntimeCall("lookupPanic", nil, "")
	b.CreateUnreachable()

	// Ok: this is a valid pointer.
	b.SetInsertPointAtEnd(nextBlock)
}

// createSliceBoundsCheck emits a bounds check before a slicing operation to make
// sure it is within bounds.
//
// This function is both used for slicing a slice (low and high have their
// normal meaning) and for creating a new slice, where 'capacity' means the
// biggest possible slice capacity, 'low' means len and 'high' means cap. The
// logic is the same in both cases.
func (b *builder) createSliceBoundsCheck(capacity, low, high, max llvm.Value, lowType, highType, maxType *types.Basic) {
	if b.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Extend the capacity integer to be at least as wide as low and high.
	capacityType := capacity.Type()
	if low.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = low.Type()
	}
	if high.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = high.Type()
	}
	if max.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = max.Type()
	}
	if capacityType != capacity.Type() {
		capacity = b.CreateZExt(capacity, capacityType, "")
	}

	// Extend low and high to be the same size as capacity.
	if low.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if lowType.Info()&types.IsUnsigned != 0 {
			low = b.CreateZExt(low, capacityType, "")
		} else {
			low = b.CreateSExt(low, capacityType, "")
		}
	}
	if high.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if highType.Info()&types.IsUnsigned != 0 {
			high = b.CreateZExt(high, capacityType, "")
		} else {
			high = b.CreateSExt(high, capacityType, "")
		}
	}
	if max.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if maxType.Info()&types.IsUnsigned != 0 {
			max = b.CreateZExt(max, capacityType, "")
		} else {
			max = b.CreateSExt(max, capacityType, "")
		}
	}

	faultBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, "slice.outofbounds")
	nextBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, "slice.next")
	b.blockExits[b.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Now do the bounds check: low > high || high > capacity
	outOfBounds1 := b.CreateICmp(llvm.IntUGT, low, high, "slice.lowhigh")
	outOfBounds2 := b.CreateICmp(llvm.IntUGT, high, max, "slice.highmax")
	outOfBounds3 := b.CreateICmp(llvm.IntUGT, max, capacity, "slice.maxcap")
	outOfBounds := b.CreateOr(outOfBounds1, outOfBounds2, "slice.lowmax")
	outOfBounds = b.CreateOr(outOfBounds, outOfBounds3, "slice.lowcap")
	b.CreateCondBr(outOfBounds, faultBlock, nextBlock)

	// Fail: this is a nil pointer, exit with a panic.
	b.SetInsertPointAtEnd(faultBlock)
	b.createRuntimeCall("slicePanic", nil, "")
	b.CreateUnreachable()

	// Ok: this is a valid pointer.
	b.SetInsertPointAtEnd(nextBlock)
}

// createNilCheck checks whether the given pointer is nil, and panics if it is.
// It has no effect in well-behaved programs, but makes sure no uncaught nil
// pointer dereferences exist in valid Go code.
func (b *builder) createNilCheck(ptr llvm.Value, blockPrefix string) {
	// Check whether we need to emit this check at all.
	if !ptr.IsAGlobalValue().IsNil() {
		return
	}

	// Check whether this is a nil pointer.
	faultBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, blockPrefix+".nil")
	nextBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, blockPrefix+".next")
	b.blockExits[b.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Compare against nil.
	var isnil llvm.Value
	if ptr.Type().PointerAddressSpace() == 0 {
		// Do the nil check using the isnil builtin, which marks the parameter
		// as nocapture.
		// The reason it has to go through a builtin, is that a regular icmp
		// instruction may capture the pointer in LLVM semantics, see
		// https://reviews.llvm.org/D60047 for details. Pointer capturing
		// unfortunately breaks escape analysis, so we use this trick to let the
		// functionattr pass know that this pointer doesn't really escape.
		ptr = b.CreateBitCast(ptr, b.i8ptrType, "")
		isnil = b.createRuntimeCall("isnil", []llvm.Value{ptr}, "")
	} else {
		// Do the nil check using a regular icmp. This can happen with function
		// pointers on AVR, which don't benefit from escape analysis anyway.
		nilptr := llvm.ConstPointerNull(ptr.Type())
		isnil = b.CreateICmp(llvm.IntEQ, ptr, nilptr, "")
	}
	b.CreateCondBr(isnil, faultBlock, nextBlock)

	// Fail: this is a nil pointer, exit with a panic.
	b.SetInsertPointAtEnd(faultBlock)
	b.createRuntimeCall("nilPanic", nil, "")
	b.CreateUnreachable()

	// Ok: this is a valid pointer.
	b.SetInsertPointAtEnd(nextBlock)
}
