//
// machine.go
// Copyright (C) 2020 Toran Sahu <toran.sahu@yahoo.com>
//
// Distributed under terms of the MIT license.
//

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	lspInternal "github.com/lighttiger2505/sqls/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/toransahu/grpc-eg-go/machine"
	"github.com/toransahu/grpc-eg-go/utils"
	"github.com/toransahu/grpc-eg-go/utils/stack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OperatorType string

const (
	PUSH OperatorType = "PUSH"
	POP               = "POP"
	ADD               = "ADD"
	SUB               = "SUB"
	MUL               = "MUL"
	DIV               = "DIV"
	FIB               = "FIB"
)

type MachineServer struct{}
type LspServer struct {
	LSP *jsonrpc2.Conn
}

// Execute runs the set of instructions given.
func (s *MachineServer) Execute(stream machine.Machine_ExecuteServer) error {
	var stack stack.Stack
	for {
		instruction, err := stream.Recv()
		if err == io.EOF {
			log.Println("EOF")
			return nil
		}
		if err != nil {
			return err
		}

		operand := instruction.GetOperand()
		operator := instruction.GetOperator()
		op_type := OperatorType(operator)

		fmt.Printf("Operand: %v, Operator: %v\n", operand, operator)

		switch op_type {
		case PUSH:
			stack.Push(float32(operand))
		case POP:
			stack.Pop()
		case ADD, SUB, MUL, DIV:
			item2, popped := stack.Pop()
			item1, popped := stack.Pop()

			if !popped {
				return status.Error(codes.Aborted, "Invalid sets of instructions. Execution aborted")
			}

			var res float32
			if op_type == ADD {
				res = item1 + item2
			} else if op_type == SUB {
				res = item1 - item2
			} else if op_type == MUL {
				res = item1 * item2
			} else if op_type == DIV {
				res = item1 / item2
			}

			stack.Push(res)
			if err := stream.Send(&machine.Result{Output: float32(res)}); err != nil {
				return err
			}
		case FIB:
			n, popped := stack.Pop()

			if !popped {
				return status.Error(codes.Aborted, "Invalid sets of instructions. Execution aborted")
			}

			if op_type == FIB {
				for f := range utils.FibonacciRange(int(n)) {
					if err := stream.Send(&machine.Result{Output: float32(f)}); err != nil {
						return err
					}
				}
			}
		default:
			return status.Errorf(codes.Unimplemented, "Operation '%s' not implemented yet", operator)
		}

	}
}

func didChange(lsp *jsonrpc2.Conn, newText string) {
	var resp interface{}

	didchangeParams := lspInternal.DidChangeTextDocumentParams{
		TextDocument: lspInternal.VersionedTextDocumentIdentifier{
			Version: 2,
			URI:     "test.sql",
		},
		ContentChanges: []lspInternal.TextDocumentContentChangeEvent{
			{Text: newText},
		},
	}

	err := lsp.Call(context.Background(), "textDocument/didChange", didchangeParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}
}

func (l *LspServer) DidChange(stream machine.Lsp_DidChangeServer) error {

	didChangeParams, err := stream.Recv()
	if err == io.EOF {
		log.Println("EOF")
		return nil
	}
	if err != nil {
		return err
	}

	changes := didChangeParams.GetContentChanges()
	updatedText := changes[0].GetText()
	didChange(l.LSP, updatedText)

	fmt.Printf("changes: %v, updatedText: %v\n", changes, updatedText)

	if err := stream.Send(&machine.Result{Output: float32(1)}); err != nil {
		return err
	}

	return nil
}

func completion(lsp *jsonrpc2.Conn, position lspInternal.Position) []lspInternal.CompletionItem {
	var resp []lspInternal.CompletionItem

	completionParams := lspInternal.CompletionParams{TextDocumentPositionParams: lspInternal.TextDocumentPositionParams{
		TextDocument: lspInternal.TextDocumentIdentifier{
			URI: "test.sql",
		},
		Position: position,
	}}

	err := lsp.Call(context.Background(), "textDocument/completion", completionParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}

	// add proper return type
	return resp
}

func (l *LspServer) Completion(stream machine.Lsp_CompletionServer) error {
	params, err := stream.Recv()

	if err != nil {
		return err
	}

	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var completionParams lspInternal.CompletionParams
	err = json.Unmarshal(b, &completionParams)

	fmt.Println("completionParams: ", completionParams)
	if err != nil {
		return err
	}

	completions := completion(l.LSP, completionParams.Position)
	// convert completions to machine.CompletionResponse

	var machineCompletions []*machine.CompletionItem
	b, err = json.Marshal(completions)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &machineCompletions)

	//item := &machine.CompletionItem{Label: "Hello world"}
	//items := []*machine.CompletionItem{item}
	res := machine.CompletionResponse{Items: machineCompletions}
	if err != nil {
		return err
	}

	if err := stream.Send(&res); err != nil {
		return err
	}

	return nil

}
