Renamer.go
==========

A batch renaming script in Go. Similar in features and scope as 
[Renamer](https://github.com/Donearm/Renamer) but in Go. It was born out of 
experimenting with Go and building some muscle memory in it. I hope it's useful 
for you (if anything, as an example)

Use
===

The flags are:

* **-p/-prefix**		Add a prefix to the beginning of filenames
* **-s/-suffix**		Add a suffix to the end of filenames
* **-i/-index**		Rename every filenames with this index string to a 
  pattern of "<index><num>.<ext>
* **-I/-startnumber**	The number to start the incremental renaming of 
  filenames according to `--index` above. It has no significance without 
  `--index`
* **-e/-lower-extension**	Lowercase the extensions
* **-l/-lowercase**	Lowercase the filenames
* **-u/-uppercase**	Uppercase the filenames
* **-x/-regexp**		Rename only files matching this regexp. If none, rename 
  all files in the target directory
* **-t/-target-dir**	The directory where are the files to rename. Default is 
  the current directory
* **-n/-dry-run**		List operations but don't actually copy/rename anything
* **-c/-copy**			Copy instead of renaming
* **-f/-force**			Overwrite existing files
* **-r/-recursive**		Operate recursively into subdirectories

At least one between `prefix`, `suffix`, `index`, `lower-extension`, `lowercase` 
and `uppercase` must be given. If no target directory specified, Renamer.go will 
operate on the current directory.

License
======

see LICENSE
