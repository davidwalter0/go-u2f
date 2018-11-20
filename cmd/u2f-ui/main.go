package main

import (
	"log"
	"os"

	"github.com/davidwalter0/go-u2f/cfg"
	"github.com/davidwalter0/go-u2f/u2f"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func init() {
	if err := cfg.Setup(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	gtk.Init(&os.Args)

	// Declarations
	Window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	RootBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	TreeView, _ = gtk.TreeViewNew()
	Status, _ = gtk.EntryNew()
	ListStore, _ = gtk.ListStoreNew(glib.TYPE_STRING)

	// Window properties
	log.Println(u2f.UnAuthenticatedTitle)
	Window.SetTitle(u2f.UnAuthenticatedTitle)
	Window.Connect("destroy", gtk.MainQuit)

	// TreeView properties
	renderer, _ := gtk.CellRendererTextNew()
	Column, _ = gtk.TreeViewColumnNewWithAttribute("", renderer, "text", 0)
	TreeView.AppendColumn(Column)
	TreeView.SetModel(ListStore)

	// TreeView selection properties
	TreeView.SetActivateOnSingleClick(true)
	TreeView.SetActivateOnSingleClick(true)

	sel, _ := TreeView.GetSelection()
	sel.SetMode(gtk.SELECTION_SINGLE)
	sel.Connect("changed", Finalize)
	sel.Connect("changed", SelectionChanged)
	sel.Connect("changed", ShowEntry)

	Status.SetEditable(false)

	RootBox.PackStart(TreeView, true, true, 0)
	RootBox.PackStart(Status, false, false, 0)

	Window.Add(RootBox)
	// Populating list
	AppendMultipleToList(
		"",
		"Register",
		"Authenticate",
	)
	LazyUpdate()
	Window.SetDefaultSize(400, 300)
	Window.SetDefaultSize(400, 150)
	Window.SetModal(true)
	Window.SetIconFromFile("028-key-card.png")
	Window.ShowAll()
	gtk.Main()
}

func StatusSetText(status *gtk.Entry, text string) bool {
	status.SetText(text)
	// Returning false here is unnecessary, as anything but returning true
	// will remove the function from being called by the GTK main loop.
	return false
}

func LazyUpdate() {
	go func() {
		var err error
		for {
			select {
			case text, ok := <-Message:
				if ok {
					// mutex.Lock()
					if cfg.Env.Debugging {
						log.Println("Before IdleAdd Status", text)
					}
					_, err = glib.IdleAdd(Status.SetText, text)
					if err != nil {
						log.Println("IdleAdd() failed", err)
					}
					if cfg.Env.Debugging {
						log.Println("After IdleAdd Status", text)
					}
					// mutex.Unlock()
				} else {
					log.Println("Messages channel is closed")
				}
			}
			// time.Sleep(time.Second)
		}
	}()

}
