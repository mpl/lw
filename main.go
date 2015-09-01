// 2015 - Mathieu Lonjaret

// The acmetags program prints the tags of the acme windows.
package main

import (
	"bytes"
//	"errors"
	"flag"
	"fmt"
//	"io/ioutil"
	"log"
	"os"
//	"os/exec"
//	"strings"
	"strconv"
	"sort"
	"time"

	"9fans.net/go/acme"
)

var (
	output    = flag.String("o", "", "output file. will only truncate if no error and output is non empty.")
	timestamp = flag.Bool("ts", false, "add a timestamp suffix to the output file name")
	allTags   = flag.Bool("all", false, "print tags of all windows, instead of only \"win\" windows.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: acmetags [-all]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

type winInfo struct {
	w acme.WinInfo
	dirty bool
	modTime time.Time // if zero -> not a file
}

type winInfos []winInfo

func (w winInfos) Len() int           { return len(w) }
func (w winInfos) Swap(i, j int)      { w[i], w[j] = w[j], w[i] }
func (w winInfos) Less(i, j int) bool {
	if w[i].dirty {
		if !w[j].dirty {
			return true
		}
		if w[i].modTime.IsZero() {
			return true
		}
		if w[j].modTime.IsZero() {
			return false
		}
		return w[i].modTime.After(w[j].modTime)	
	}
	if w[j].dirty {
		return false
	}
	return w[i].modTime.After(w[j].modTime)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var allWins []winInfo

	windows, err := acme.Windows()
	if err != nil {
		log.Fatalf("could not get acme windows: %v", err)
	}
	for _, win := range windows {
		w, err := acme.Open(win.ID, nil)
		if err != nil {
			log.Fatalf("could not open window (%v, %d): %v", win.Name, win.ID, err)
		}
		defer w.CloseFiles()
		b, err := w.ReadAll("ctl")
		if err != nil {
			log.Fatalf("could not read ctl file of (%v, %d): %v", win.Name, win.ID, err)
		}
		fields := bytes.Fields(b)
		if len(fields) != 8 {
			log.Fatalf("unexpected number of fields for (%v, %d): wanted %v, got %v", win.Name, win.ID, 8, len(fields))
		}
		isDirty, _ := strconv.ParseBool(string(fields[4]))
		wini := winInfo {
			w: win,
			dirty: isDirty,
		}
		fi, err := os.Stat(win.Name)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("could not stat disk file of (%v, %d): %v", win.Name, win.ID, err)
			}
		} else {
			wini.modTime = fi.ModTime()
		}
		allWins = append(allWins, wini)
	}
	sort.Sort(winInfos(allWins))
	for _,v := range allWins {
		fmt.Printf("%v	%v\n", v.dirty, v.w.Name)
	}
}
