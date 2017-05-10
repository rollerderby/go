#!/bin/bash

PRINTFUNCS=Print,Printf,Println,Fprint,Fprintf,Fprintln,Sprint,Sprintf,Sprintln,Error,Errorf,Fatal,Fatalf,Log,Logf,Panic,Panicf,Panicln,Emerg,Fatal,Alert,Crit,Err,Error,Warning,Notice,Info,Debug,Emergf,Fatalf,Alertf,Critf,Errf,Errorf,Warningf,Noticef,Infof,Debugf
VET="go tool vet -all -printfuncs $PRINTFUNCS"
BASE=github.com/rollerderby/go
PACKAGE=$BASE/cmd/server

cd `go env GOPATH`/src
pwd
$VET $PACKAGE
go list -f '{{.Deps}}' $PACKAGE | tr "[" " " | tr "]" " " | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | grep $BASE | xargs -n 1 $VET
