/*
renamer.go

Renaming files script on various patterns

Author: Gianluca Fiore <forod.g@gmail.com> © 2013-2014
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var usageMessage string = `
renamer.go (-p <prefix>|-s <suffix>|-i <indexname> -I <num>|-e|-l|-u) [-x <regexp>] [-t <target_dir>] -[cnr]

renamer.go will rename all files matching a regexp or all files in the 
given directory (and optionally in all its subdirectories) by the flag 
chosen. It can add a prefix/suffix, rename according to a 
<name><numeric_index> pattern, make all low/uppercase, lowercase the 
extension and copy files somewhere else leaving the originals untouched.

Arguments:

	-prefix|-p <prefix>
		Renames matching files by adding a prefix

	-suffix|-s <suffix>
		Renames matching files by adding a suffix

	-index|-i <name>
		Renames matching files to <name><num>. Requires '-I'

	-startnumber|-I <num>
		The <num> to start renaming in index mode. Requires '-i'

	-regexp|-x <regexp>
		The regular expression for matching files. Use double quotes 
		around it

	-target-dir|-t <path>
		The directory where to rename/copy the files. Default is the 
		current directory

	-copy|-c
		Copy instead of renaming

	-lower-extension|-e
		Lowercase the extension

	-lowercase|-l
		Lowercase the new filename. It is mutually exclusive with 
		'-u'

	-uppercase|-u
		Uppercase the new filename. It is mutually exclusive with 
		'-l'

	-dry-run|-n
		List files but don't actually perform any action

	-force|-f
		Overwrite existing files. The default is to not copy/rename 
		if the target file already exists

	-recursive|-r
		Operate recursively on all subdirectories of target-dir
`

var regexpArg string         // the regexp argument
var fileRegex *regexp.Regexp // the files matching the regexp
var prefixArg string         // the prefix
var suffixArg string         // the suffix
var indexArg string          // the <name> in the <name><num> pattern
var numArg int               // the <num> in the <name><num> pattern
var targetArg string         // the target directory
var lowerExtArg bool         // the lower extension switch
var lowerArg bool            // the lowercase switch
var upperArg bool            // the uppercase switch
var copyArg bool             // the copy switch
var dryrunArg bool           // the dry-run switch
var forceArg bool            // the force switch
var recursiveArg bool        // the recursive switch

var operationSuccessful int // numeric flag to keep trace of what went
// wrong during the renaming

// Print a message and the usage instructions
func printUsage(msg string) {
	fmt.Fprintf(os.Stderr, msg+"\n")
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

// Parse flags ("arguments")
func flagsInit() {
	// Use our custom usageMessage to print usage instructions
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMessage)
	}

	const (
		def_regexp    = ""
		def_prefix    = ""
		def_suffix    = ""
		def_index     = ""
		def_num       = 1
		def_target    = "."
		def_lowerext  = false
		def_lowercase = false
		def_uppercase = false
		def_copy      = false
		def_dryrun    = false
		def_force     = false
		def_recursive = false
	)

	flag.StringVar(&regexpArg, "regexp", def_regexp, "")
	flag.StringVar(&regexpArg, "x", def_regexp, "")
	flag.StringVar(&prefixArg, "prefix", def_prefix, "")
	flag.StringVar(&prefixArg, "p", def_prefix, "")
	flag.StringVar(&suffixArg, "suffix", def_suffix, "")
	flag.StringVar(&suffixArg, "s", def_suffix, "")
	flag.StringVar(&indexArg, "index", def_index, "")
	flag.StringVar(&indexArg, "i", def_index, "")
	flag.IntVar(&numArg, "startnumber", def_num, "")
	flag.IntVar(&numArg, "I", def_num, "")
	flag.StringVar(&targetArg, "target", def_target, "")
	flag.StringVar(&targetArg, "t", def_target, "")
	flag.BoolVar(&lowerExtArg, "lower-extension", def_lowerext, "")
	flag.BoolVar(&lowerExtArg, "e", def_lowerext, "")
	flag.BoolVar(&lowerArg, "lowercase", def_lowercase, "")
	flag.BoolVar(&lowerArg, "l", def_lowercase, "")
	flag.BoolVar(&upperArg, "uppercase", def_uppercase, "")
	flag.BoolVar(&upperArg, "u", def_uppercase, "")
	flag.BoolVar(&copyArg, "copy", def_copy, "")
	flag.BoolVar(&copyArg, "c", def_copy, "")
	flag.BoolVar(&dryrunArg, "dry-run", def_dryrun, "")
	flag.BoolVar(&dryrunArg, "n", def_dryrun, "")
	flag.BoolVar(&forceArg, "force", def_force, "")
	flag.BoolVar(&forceArg, "f", def_force, "")
	flag.BoolVar(&recursiveArg, "recursive", def_recursive, "")
	flag.BoolVar(&recursiveArg, "r", def_recursive, "")

	flag.Parse()

	if regexpArg == "" && prefixArg == "" && suffixArg == "" && indexArg == "" && lowerExtArg == false && lowerArg == false && upperArg == false {
		printUsage("At least one of the mandatory actions must be given, nothing to do...")
	}
}

// Write a renamed or a copy of a file to disk
func writeFile(oldname, newname string) {
	// check if the new filename is already present
	_, lstat_err := os.Lstat(newname)
	if lstat_err == nil && forceArg == false {
		fmt.Fprintf(os.Stderr, "File %s already exist! Use -force to override it\n", newname)
		operationSuccessful = operationSuccessful + 1
	}
	if dryrunArg {
		// if dry-run was given, just output the renaming operation
		fmt.Fprintf(os.Stdout, "Renaming %s to %s (dry-run)\n", oldname, newname)
		operationSuccessful = operationSuccessful + 0
	} else {
		if copyArg {
			copyf, create_err := os.Create(newname)
			if create_err != nil {
				fmt.Fprintf(os.Stderr, create_err.Error())
				operationSuccessful = operationSuccessful + 1
			}
			originalf, open_err := os.Open(oldname)
			if open_err != nil {
				fmt.Fprintf(os.Stderr, open_err.Error())
				operationSuccessful = operationSuccessful + 1
			}
			_, copy_err := io.Copy(copyf, originalf)
			if copy_err != nil {
				fmt.Fprintf(os.Stderr, "An error occurred during the copy of %s to %s\n", oldname, newname)
				fmt.Fprintf(os.Stderr, copy_err.Error())
				operationSuccessful = operationSuccessful + 1
			} else {
				fmt.Fprintf(os.Stdout, "Copying %s to %s\n", oldname, newname)
			}
		} else {
			rename_err := os.Rename(oldname, newname)
			if rename_err != nil {
				fmt.Fprintf(os.Stderr, "An error occurred during the renaming of %s to %s\n", oldname, newname)
				fmt.Fprintf(os.Stderr, rename_err.Error())
				operationSuccessful = operationSuccessful + 1
			} else {
				fmt.Fprintf(os.Stdout, "Renaming %s to %s\n", oldname, newname)
			}
		}
	}
}

// Add a prefix string to a name
func addPrefix(names []string, prefix string) int {
	var finalname, dirname string
	for _, f := range names {
		dirname = filepath.Dir(f)
		finalname = filepath.Join(dirname, prefix+filepath.Base(f))
		writeFile(f, finalname)
	}
	return 0
}

// Add a suffix string to a name
func addSuffix(names []string, suffix string) int {
	var finalname, dirname, justname, ext string
	for _, f := range names {
		ext = filepath.Ext(f)
		dirname = filepath.Dir(f)
		justname = strings.TrimSuffix(filepath.Base(f), ext)
		finalname = filepath.Join(dirname, justname+suffix+ext)
		writeFile(f, finalname)
	}
	return 0
}

// Rename a slice of filenames to <newname><count>.<extension>
func indexName(names []string, newname string, count int) int {
	var finalname, dirname, ext string
	for _, f := range names {
		ext = filepath.Ext(f)
		dirname = filepath.Dir(f)
		finalname = fmt.Sprintf("%s/%s%03d%s", dirname, newname, count, ext)
		writeFile(f, finalname)
		count++
	}
	return 0
}

// Make extensions lowercase
func lowercaseExtension(names []string) int {
	var finalname, dirname, basename, ext string
	for _, f := range names {
		dirname = filepath.Dir(f)
		basename = filepath.Base(f)
		ext = filepath.Ext(f)
		finalname = filepath.Join(dirname, strings.TrimSuffix(basename, ext)+strings.ToLower(strings.TrimSuffix(ext, basename)))
	}
	return 0
}

// Make filenames all lowercase
func lowercaseFiles(names []string) int {
	var finalname, dirname string
	for _, f := range names {
		dirname = filepath.Dir(f)
		finalname = filepath.Join(dirname, strings.ToLower(filepath.Base(f)))
		writeFile(f, finalname)
	}
	return 0
}

// Make filenames all uppercase
func uppercaseFiles(names []string) int {
	var finalname, dirname string
	for _, f := range names {
		dirname = filepath.Dir(f)
		finalname = filepath.Join(dirname, strings.ToUpper(filepath.Base(f)))
		writeFile(f, finalname)
	}
	return 0
}

// Get all files and directories
func getFilesFromDir(dirname string) ([]string, []string) {
	var complete_path string                // final, absolute, path
	var filesindir = make([]os.FileInfo, 0) // files & directories found in path
	var allfiles = make([]string, 0)
	var alldirectories = make([]string, 0)

	dirinfo, lerr := os.Lstat(dirname)
	if lerr != nil {
		fmt.Fprintf(os.Stderr, lerr.Error())
		return alldirectories, allfiles
	}

	// check whether targetArg is an absolute path AND a directory
	if filepath.IsAbs(dirname) && dirinfo.IsDir() {
		complete_path = dirname
	} else {
		abs_path, err := filepath.Abs(dirname)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
		complete_path = abs_path
	}
	dir, err := os.Open(complete_path)
	defer dir.Close()
	if err != nil {
		err = errors.New(fmt.Sprintf("Target directory %s is not a directory or can't be accessed\n", complete_path))
		fmt.Fprintf(os.Stderr, err.Error())
		return alldirectories, allfiles
	}

	// scan for files/directories in path
	filesindir, read_err := dir.Readdir(0)
	if read_err != nil {
		fmt.Fprintf(os.Stderr, read_err.Error())
		return alldirectories, allfiles
	}

	// check in filesindir slice and separate directories from files in
	// 2 different slices
	for _, f := range filesindir {
		if f.IsDir() {
			alldirectories = append(alldirectories, filepath.Join(complete_path, f.Name()))
		} else {
			allfiles = append(allfiles, filepath.Join(complete_path, f.Name()))
		}
	}

	return alldirectories, allfiles
}

func renameFiles(dir, files []string) int {
	var basename string
	var matchingfiles []string // a slice containing only the files
	// matching the regexp passed as argument (if)
	var result int // the integer returned by each functions,
	// signaling success or failure

	// recursively search on every directory in dir for other
	// files/directories if recursiveArg switch has been enabled
	if dir != nil && recursiveArg == true {
		for _, d := range dir {
			nd, nf := getFilesFromDir(d)
			// if it's a dir, append to []dir
			if len(nd) > 0 {
				for _, i := range nd {
					dir = append(dir, i)
				}
			}
			// if it's a file, append to []files
			if len(nf) > 0 {
				for _, i := range nf {
					files = append(files, i)
				}
			}
		}
	}

	// check if the files should match a given regexp
	if regexpArg != "" {
		for _, f := range files {
			basename = filepath.Base(f)
			// Compile and check if the regexp is a valid one
			compRegexp, err := regexp.Compile(regexpArg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid regexp: %s\n", regexpArg)
				printUsage("You must give a valid regexp (or none, to operate on all files). Alternatively, add -force to force renaming all files, whether they match the regexp or not")
				return 1
			} else {
				if compRegexp.MatchString(basename) == false {
					continue
				} else {
					matchingfiles = append(matchingfiles, f)
				}
			}
		}
	}

	// if matchingfiles contains something, iterate over it.
	// Otherwise, we simply use the generic files slice
	if len(matchingfiles) >= 1 {
		files = matchingfiles
	} else {
		// different messages whether we have forceArg == true or false
		if regexpArg != "" {
			if forceArg {
				fmt.Fprintf(os.Stdout, "No files matched, including all files anyway (-force enabled)\n")
			} else {
				fmt.Fprintf(os.Stderr, "No files matched, check the correctness of your regexp\n")
				return 1
			}
		}
	}

	if prefixArg != "" {
		result = addPrefix(files, prefixArg)
	}
	if suffixArg != "" {
		result = addSuffix(files, suffixArg)
	}
	if indexArg != "" {
		result = indexName(files, indexArg, numArg)
	}
	if lowerArg == true && upperArg == true {
		// can't use both
		printUsage("Can't use both lowercase and uppercase, choose one only!")
		return 1
	}
	if lowerExtArg == true {
		result = lowercaseExtension(files)
	}
	if lowerArg == true {
		result = lowercaseFiles(files)
	}
	if upperArg == true {
		result = uppercaseFiles(files)
	}

	return result
}

func main() {
	var success_rename int
	var directories, files []string

	flagsInit()

	directories, files = getFilesFromDir(targetArg)

	success_rename = renameFiles(directories, files)

	// check that everything went smoothly
	if success_rename == 0 && operationSuccessful == 0 {
		fmt.Fprintf(os.Stdout, "\nRenaming complete\n")
	} else {
		fmt.Fprintf(os.Stdout, "\nNot all files were correctly renamed, check the previous error messages")
	}
}
