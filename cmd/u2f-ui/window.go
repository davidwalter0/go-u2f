package main

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/davidwalter0/go-u2f/cfg"
	"github.com/davidwalter0/go-u2f/u2f"
	"github.com/gotk3/gotk3/gtk"
)

const (
	Registered             = "Registered"
	Authenticated          = "Authenticated"
	AuthenticationFailed   = "Authentication: Finalize"
	MissingKey             = "Missing Key"
	RegistrationFailed     = "Registration: Finalize"
	PressKeyToAuthenticate = "Press key to authenticate"
)

var (
	Message                  chan string = make(chan string, 1)
	PrimaryTitle                         = "%20.20s Security Key U2F -- %s"
	UnAuthenticatedTitle                 = fmt.Sprintf(PrimaryTitle, " ", "UnAuthenticated")
	AuthenticatedTitle                   = fmt.Sprintf(PrimaryTitle, " ", "Authenticated")
	RegisteredTitle                      = fmt.Sprintf(PrimaryTitle, " ", "Registered")
	InsertSecurityKeyMessage             = "Insert Security Key"
	Window                   *gtk.Window
	RootBox                  *gtk.Box
	TreeView                 *gtk.TreeView
	Status                   *gtk.Entry
	ListStore                *gtk.ListStore
	Column                   *gtk.TreeViewColumn
	Action                   string
	LastAction               string
	BottomLabel              *gtk.Label
)

// Appends single value to the TreeView's model
func AppendToList(value string) {
	ListStore.SetValue(ListStore.Append(), 0, value)
}

// Appends several values to the TreeView's model
func AppendMultipleToList(values ...string) {
	for _, v := range values {
		AppendToList(v)
	}
}

func GetTreeSelectionName(tree *gtk.TreeSelection) string {
	rows := tree.GetSelectedRows(ListStore)
	var name string
	var path *gtk.TreePath
	for l := rows; l != nil; l = l.Next() {
		path = l.Data().(*gtk.TreePath)
		iter, _ := ListStore.GetIter(path)
		value, _ := ListStore.GetValue(iter, 0)
		name, _ = value.GetString()
		return name
	}
	return ""
}

var Semaphore = new(int64)

func SelectionChanged(tree *gtk.TreeSelection) {
	called := atomic.AddInt64(Semaphore, 1)
	defer cfg.Env.Trace("SelectionChanged")()
	name := GetTreeSelectionName(tree)
	//	once.Do(func() { Deselect(tree) })
	switch called {
	case 1:
		Deselect(tree)
	default:
	}
	Action = name
}

func ShowEntry(tree *gtk.TreeSelection) {
	switch Action {
	case "Register", "Authenticate":
		// mutex.Lock()
		Message <- fmt.Sprintf("%s: %s", Action, PressKeyToAuthenticate)
		// mutex.Unlock()
		if err := u2f.U2FAction(Action, Message); err != nil {
			Column.SetTitle(err.Error())
		}
	case "Registered":
	case "Authenticated":
	default:
	}
}

func Deselect(tree *gtk.TreeSelection) {
	var rows = tree.GetSelectedRows(ListStore)
	var path *gtk.TreePath
	for l := rows; l != nil; l = l.Next() {
		path = l.Data().(*gtk.TreePath)
	}
	if path != nil {
		tree.UnselectPath(path)
	}
}

// Handler of "changed" signal of TreeView's selection
func Finalize(s *gtk.TreeSelection) {
	defer cfg.Env.Trace("Finalize")()
	switch Action {
	case Registered, Authenticated:
	case MissingKey, RegistrationFailed, AuthenticationFailed:
		Deselect(s)
	}
}

func Act(entry *gtk.Entry) {
	defer cfg.Env.Trace("Act")()
	// if IgnAct() {
	// 	return
	// }
	// LastAction = Action
	// U2FAction(LastAction)
}

func IgnAct() bool {
	defer cfg.Env.Trace("IgnAct")()

	if LastAction == Action {
		return true
	}

	switch Action {
	case
		Registered,
		Authenticated,
		RegisteredTitle,
		AuthenticatedTitle,
		UnAuthenticatedTitle,
		MissingKey,
		AuthenticationFailed,
		RegistrationFailed:
		return true
	default:
	}
	return false
}

func ResetTitle(title string) {
	log.Println("ResetTitle", title)
	Window.SetTitle(title)
}
