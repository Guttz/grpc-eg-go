//
// machine.go
// Copyright (C) 2020 Toran Sahu <toran.sahu@yahoo.com>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"github.com/toransahu/grpc-eg-go/machine"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:9111", "The server address in the format of host:port")
	lspClient  machine.LspClient
)

func runExecute(client machine.MachineClient, instructions []*machine.Instruction) {
	log.Printf("Streaming %v", instructions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Execute(ctx)
	if err != nil {
		log.Fatalf("%v.Execute(ctx) = %v, %v: ", client, stream, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Err: %v", err)
			}
			log.Printf("output: %v", result.GetOutput())
		}
	}()

	for _, instruction := range instructions {
		if err := stream.Send(instruction); err != nil {
			log.Fatalf("%v.Send(%v) = %v: ", stream, instruction, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}
	<-waitc
}

func RunDidChange(client machine.LspClient, didChangeParams *machine.DidChangeTextDocumentParams) {
	log.Printf("Streaming %v", didChangeParams)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.DidChange(ctx)
	if err != nil {
		log.Fatalf("%v.DidChange(ctx) = %v, %v: ", client, stream, err)
	}
	waitc := make(chan struct{})

	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Err: %v", err)
			}
			log.Printf("output: %v", result.GetOutput())
		}
	}()

	if err := stream.Send(didChangeParams); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", stream, didChangeParams, err)
	}

	time.Sleep(500 * time.Millisecond)
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}
	<-waitc
}

func Completion(client machine.LspClient, completionParams *machine.CompletionParams) {
	log.Printf("Streaming %v", completionParams)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Completion(ctx)
	if err != nil {
		log.Fatalf("%v.DidChange(ctx) = %v, %v: ", client, stream, err)
	}
	waitc := make(chan struct{})

	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				log.Println("EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Err: %v", err)
			}

			log.Printf("completions: %v", result.GetItems())
		}
	}()

	if err := stream.Send(completionParams); err != nil {
		log.Fatalf("%v.Send(%v) = %v: ", stream, completionParams, err)
	}

	time.Sleep(50 * time.Millisecond)
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got error %v, want %v", stream, err, nil)
	}
	<-waitc
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := machine.NewMachineClient(conn)
	lspClient = machine.NewLspClient(conn)

	t := &machine.TextDocumentContentChangeEvent{
		Text: "se",
	}
	tt := []*machine.TextDocumentContentChangeEvent{t}

	RunDidChange(lspClient, &machine.DidChangeTextDocumentParams{
		TextDocument: &machine.VersionedTextDocumentIdentifier{
			Version: 2,
			URI:     "test.sql",
		},
		ContentChanges: tt,
	})

	completionParams := &machine.CompletionParams{
		TextDocumentPositionParams: &machine.TextDocumentPositionParams{
			TextDocument: &machine.TextDocumentIdentifier{
				URI: "test.sql",
			},
			Position: &machine.Position{
				Line:      0,
				Character: 2,
			},
		}}

	Completion(lspClient, completionParams)

	// try Execute()
	instructions := []*machine.Instruction{
		{Operand: 1, Operator: "PUSH"},
		{Operand: 2, Operator: "PUSH"},
		{Operator: "ADD"},
	}

	runExecute(client, instructions)
}
