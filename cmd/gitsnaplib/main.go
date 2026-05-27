package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"unsafe"

	"github.com/spmurray/gitsnap/internal/gitsnap"
)

type response struct {
	OK    bool   `json:"ok"`
	Value any    `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

var client = gitsnap.New()

//export GitsnapInit
func GitsnapInit(worktree *C.char) *C.char {
	return result(nil, client.Init(context.Background(), worktreeStr(worktree)))
}

//export GitsnapCleanup
func GitsnapCleanup(worktree *C.char) *C.char {
	return result(nil, client.Cleanup(worktreeStr(worktree)))
}

//export GitsnapSave
func GitsnapSave(worktree *C.char, name *C.char) *C.char {
	value, err := client.Save(
		context.Background(),
		worktreeStr(worktree),
		cstr(name),
	)
	return result(value, err)
}

//export GitsnapResolve
func GitsnapResolve(worktree *C.char, name *C.char) *C.char {
	value, err := client.Resolve(worktreeStr(worktree), cstr(name))
	return result(value, err)
}

//export GitsnapDiff
func GitsnapDiff(worktree *C.char, name *C.char) *C.char {
	value, err := client.Diff(context.Background(), worktreeStr(worktree), cstr(name))
	return result(value, err)
}

//export GitsnapFiles
func GitsnapFiles(worktree *C.char, name *C.char) *C.char {
	value, err := client.Files(
		context.Background(),
		worktreeStr(worktree),
		cstr(name),
	)
	return result(value, err)
}

//export GitsnapRestore
func GitsnapRestore(
	worktree *C.char,
	name *C.char,
	pathsJSON *C.char,
) *C.char {
	paths, err := paths(cstr(pathsJSON))
	if err != nil {
		return result(nil, err)
	}
	err = client.Restore(
		context.Background(),
		worktreeStr(worktree),
		cstr(name),
		paths,
	)
	return result(nil, err)
}

//export GitsnapAliases
func GitsnapAliases(worktree *C.char) *C.char {
	value, err := client.Aliases(worktreeStr(worktree))
	return result(value, err)
}

//export GitsnapFree
func GitsnapFree(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

func main() {}

func worktreeStr(s *C.char) string {
	value := cstr(s)
	if value == "" {
		return "."
	}
	return value
}

func cstr(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

func paths(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	var paths []string
	if err := json.Unmarshal([]byte(s), &paths); err != nil {
		return nil, err
	}
	return paths, nil
}

func result(value any, err error) *C.char {
	res := response{OK: err == nil, Value: value}
	if err != nil {
		res.Error = err.Error()
	}
	b, marshalErr := json.Marshal(res)
	if marshalErr != nil {
		b, _ = json.Marshal(response{OK: false, Error: marshalErr.Error()})
	}
	return C.CString(string(b))
}
